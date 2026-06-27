// Package hotkey detects a tap of a single modifier key (by default the right
// Command key) using a listen-only macOS event tap. A quick standalone tap
// fires the callback; using the key as a modifier in a shortcut does not, so
// the key keeps working normally.
//
// Requires Accessibility permission (the same grant used for pasting).
package hotkey

/*
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework CoreFoundation
#include "chotkey.h"
*/
import "C"

import (
	"errors"
	"runtime"
	"strings"

	"github.com/kzkr/silentrec/internal/config"
)

// ErrNotTrusted means the process lacks Accessibility permission, so the event
// tap could not be created.
var ErrNotTrusted = errors.New("hotkey: Accessibility permission required")

// fired is the callback invoked on each detected tap. Set once in New.
var fired func()

//export silentrecHotkeyFired
func silentrecHotkeyFired() {
	if fired != nil {
		fired()
	}
}

// Hotkey owns the running event tap.
type Hotkey struct{}

// New starts watching for taps of the configured key and invokes onPress for
// each. Returns ErrNotTrusted if Accessibility has not been granted.
func New(cfg config.Hotkey, onPress func()) (*Hotkey, error) {
	keycode := keycodeFor(cfg.Key)
	fired = onPress

	errc := make(chan error, 1)
	go func() {
		// The tap must be created and run on the same OS thread.
		runtime.LockOSThread()
		if int(C.silentrec_hotkey_create(C.int(keycode))) != 0 {
			errc <- ErrNotTrusted
			return
		}
		errc <- nil
		C.silentrec_hotkey_run() // blocks until Close
	}()

	if err := <-errc; err != nil {
		return nil, err
	}
	return &Hotkey{}, nil
}

// Close stops the event tap.
func (h *Hotkey) Close() {
	C.silentrec_hotkey_stop()
}

// keycodeFor maps a config key name to a macOS virtual keycode. Defaults to the
// right Command key.
func keycodeFor(key string) int {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "left_command", "left_cmd", "lcmd":
		return 55
	case "right_option", "right_alt", "ropt":
		return 61
	case "right_control", "right_ctrl", "rctrl":
		return 62
	case "right_command", "right_cmd", "rcmd", "":
		return 54
	default:
		return 54
	}
}
