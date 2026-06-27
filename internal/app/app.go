// Package app wires the components together and owns the recording state
// machine and the UI (menu-bar item and result popup).
//
// Flow: a global hotkey toggles between idle and recording. While recording the
// menu-bar icon blinks red; on stop, audio is transcribed with whisper, lightly
// cleaned, then pasted into the frontmost app (or shown in a popup when that app
// is the Finder/desktop, where there's nowhere to paste).
package app

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"

	"github.com/kzkr/silentrec/internal/audio"
	"github.com/kzkr/silentrec/internal/cleanup"
	"github.com/kzkr/silentrec/internal/config"
	"github.com/kzkr/silentrec/internal/hotkey"
	"github.com/kzkr/silentrec/internal/paste"
	"github.com/kzkr/silentrec/internal/snippets"
	"github.com/kzkr/silentrec/internal/transcribe"
)

//go:embed tray-idle.png
var trayIdlePNG []byte

//go:embed tray-rec.png
var trayRecPNG []byte

type state int

const (
	stateIdle state = iota
	stateRecording
	stateProcessing
)

// App holds all runtime state. Construct with New, then call Run.
type App struct {
	fyne fyne.App
	desk desktop.App

	iconIdle fyne.Resource
	iconRec  fyne.Resource

	cfg      config.Config
	recorder *audio.Recorder
	hotkey   *hotkey.Hotkey

	// model is loaded asynchronously; ready is closed once loading finishes
	// (whether it succeeded or not). transcriber stays nil if loading failed.
	transcriber *transcribe.Transcriber
	ready       chan struct{}

	mu    sync.Mutex
	state state

	// hadPasteTarget records whether the frontmost app at record start can
	// receive a paste (vs. the Finder/desktop); decides paste vs. popup.
	hadPasteTarget bool

	// blinkStop stops the recording blink goroutine (guarded by mu).
	blinkStop chan struct{}
}

// New constructs the app: loads config, initializes audio, builds UI, and kicks
// off asynchronous model loading. Must be called on the main goroutine.
func New() (*App, error) {
	fa := fyneapp.NewWithID("com.kzkr.silentrec")
	desk, ok := fa.(desktop.App)
	if !ok {
		return nil, fmt.Errorf("SilentRec requires a desktop environment with a system tray")
	}
	// White windows with a black accent, matching the logo (not the dark theme).
	fa.Settings().SetTheme(silentTheme{})

	cfg, err := config.Load()
	if err != nil {
		log.Printf("config: %v (using defaults)", err)
		cfg = config.Default()
	}

	rec, err := audio.NewRecorder()
	if err != nil {
		return nil, err
	}

	a := &App{
		fyne: fa,
		desk: desk,
		// Idle icon is a template resource so macOS renders it adaptively
		// (white on a dark menu bar, dark on a light one), matching other icons.
		iconIdle: theme.NewThemedResource(fyne.NewStaticResource("tray-idle.png", trayIdlePNG)),
		// Recording icon is a literal-colour resource so the red shows as-is.
		iconRec:  fyne.NewStaticResource("tray-rec.png", trayRecPNG),
		cfg:      cfg,
		recorder: rec,
		ready:    make(chan struct{}),
		state:    stateIdle,
	}

	a.installMenu()
	a.refreshTray()

	// Load the speech model in the background (it can take several seconds).
	go a.loadModel()

	// Auto-paste and focused-field detection need Accessibility permission
	// (distinct from the Input Monitoring the hotkey tap may use). Prompt if it
	// hasn't been granted yet.
	if !paste.AccessibilityTrusted(false) {
		paste.AccessibilityTrusted(true)
		log.Printf("grant Accessibility permission to SilentRec for auto-paste, then relaunch")
	}

	hk, err := hotkey.New(cfg.Hotkey, a.onHotkey)
	if err != nil {
		if errors.Is(err, hotkey.ErrNotTrusted) {
			// Trigger the macOS Accessibility prompt; the tap works after the
			// user grants permission and relaunches SilentRec.
			paste.AccessibilityTrusted(true)
			log.Printf("hotkey: grant Accessibility permission to SilentRec, then relaunch")
		} else {
			log.Printf("hotkey: %v", err)
		}
	}
	a.hotkey = hk

	return a, nil
}

// Run starts the Fyne event loop and blocks until the app quits.
func (a *App) Run() {
	defer a.recorder.Close()
	a.fyne.Run()
}

func (a *App) loadModel() {
	defer close(a.ready)
	tr, err := transcribe.Load(a.cfg.ModelPath)
	if err != nil {
		log.Printf("model load failed: %v", err)
		return
	}
	a.transcriber = tr
	log.Printf("model loaded: %s", a.cfg.ModelPath)
}

// onHotkey is invoked from the hotkey goroutine on each tap of the hotkey.
func (a *App) onHotkey() {
	a.mu.Lock()
	switch a.state {
	case stateIdle:
		a.state = stateRecording
		a.mu.Unlock()
		a.startRecording()
	case stateRecording:
		a.state = stateProcessing
		a.mu.Unlock()
		a.refreshTray()
		go a.finish()
	default:
		a.mu.Unlock() // already processing: ignore
	}
}

func (a *App) startRecording() {
	// Capture the frontmost app on the main thread (NSWorkspace is best used
	// there; the hotkey callback runs on the event-tap thread).
	fyne.DoAndWait(func() {
		paste.RememberTarget()
		a.hadPasteTarget = paste.HasPasteTarget()
	})
	if err := a.recorder.Start(); err != nil {
		log.Printf("record start: %v", err)
		a.toIdle()
		return
	}
	a.refreshTray() // starts the red blink
}

// finish stops recording, transcribes, and delivers the text. Runs in its own
// goroutine so transcription never blocks the hotkey listener.
func (a *App) finish() {
	defer a.toIdle()

	samples, err := a.recorder.Stop()
	if err != nil {
		log.Printf("record stop: %v", err)
		return
	}

	<-a.ready // ensure the model has finished loading
	if a.transcriber == nil {
		a.showResult("Speech model failed to load. Check the model path in settings.")
		return
	}

	prompt := strings.Join(a.cfg.Vocabulary, ", ")
	text, err := a.transcriber.Transcribe(samples, a.cfg.Language, prompt)
	if err != nil {
		log.Printf("transcribe: %v", err)
		return
	}
	text = cleanup.Clean(text)
	text = snippets.Expand(text, a.cfg.Snippets)
	if text == "" {
		return
	}

	// Paste into the frontmost app; but on the Finder/desktop there's nowhere to
	// paste, so show the transcript in a popup instead of firing a Cmd+V that
	// just beeps. The transcript is left on the clipboard either way (see
	// paste.PasteText / SetClipboard).
	if a.hadPasteTarget {
		paste.PasteText(text)
		return
	}
	paste.SetClipboard(text)
	a.showResult(text)
}

func (a *App) toIdle() {
	a.mu.Lock()
	a.state = stateIdle
	a.mu.Unlock()
	a.refreshTray()
}

// --- menu-bar item + recording indicator ---

// installMenu sets the (static) menu-bar menu. Recording state is conveyed by
// the icon, not by menu text.
func (a *App) installMenu() {
	menu := fyne.NewMenu("SilentRec",
		fyne.NewMenuItem("Quit", func() { a.fyne.Quit() }),
	)
	fyne.Do(func() { a.desk.SetSystemTrayMenu(menu) })
}

// refreshTray updates the menu-bar icon for the current state: idle = adaptive
// (white/dark) S, recording = blinking red S, transcribing = solid red S.
func (a *App) refreshTray() {
	a.mu.Lock()
	st := a.state
	a.mu.Unlock()
	a.applyIcon(st)
}

func (a *App) applyIcon(st state) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stopBlinkLocked()

	switch st {
	case stateRecording:
		a.startBlinkLocked()
	case stateProcessing:
		a.setIcon(a.iconRec) // solid red while transcribing
	default:
		a.setIcon(a.iconIdle)
	}
}

func (a *App) setIcon(r fyne.Resource) {
	fyne.Do(func() { a.desk.SetSystemTrayIcon(r) })
}

// startBlinkLocked alternates the icon between the red and idle icons to blink.
// Caller holds mu.
func (a *App) startBlinkLocked() {
	stop := make(chan struct{})
	a.blinkStop = stop
	rec, idle := a.iconRec, a.iconIdle

	go func() {
		a.setIcon(rec)
		t := time.NewTicker(450 * time.Millisecond)
		defer t.Stop()
		on := true
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				on = !on
				if on {
					a.setIcon(rec)
				} else {
					a.setIcon(idle)
				}
			}
		}
	}()
}

// stopBlinkLocked halts any running blink goroutine. Caller holds mu.
func (a *App) stopBlinkLocked() {
	if a.blinkStop != nil {
		close(a.blinkStop)
		a.blinkStop = nil
	}
}
