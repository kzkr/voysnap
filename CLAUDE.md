# CLAUDE.md

Guidance for working in this repo. Read this before making changes.

## What this is

**SilentRec** — a free, fully-local, offline voice-to-text dictation menu-bar app
for macOS. Tap the **right ⌘ key** to start recording, tap it again to stop; the
audio is transcribed on-device with whisper.cpp (Metal), lightly cleaned, and
pasted into whatever text field is focused. If nothing editable is focused, the
text is shown in a popup (and left on the clipboard).

It's a personal project of Valentin's "South Computer Company" lab — a local,
owned replacement for paid tools like Superwhisper / Wispr Flow.

> Naming note: the project was renamed from "Voice" to **SilentRec**. The Go
> module is `github.com/kzkr/silentrec` and the bundle id is
> `com.kzkr.silentrec`. The **directory on disk is still `~/Desktop/voice`** —
> that's fine; don't rename it (tooling paths depend on it).

## Platform & constraints

- **macOS-only, Apple Silicon.** Heavy use of cgo (Objective-C / C) for the
  macOS-specific pieces. Not portable; don't add cross-platform abstractions.
- **Go 1.24+**, `cmake`, Xcode command-line tools.
- **Offline by design.** No network calls, no accounts. Multilingual via
  whisper auto-detect (the model is multilingual); default language is `auto`.

## Core design decisions (and why)

- **Speed-first, no LLM cleanup.** Whisper already punctuates and capitalizes.
  Post-processing is *only* whitespace trimming (`internal/cleanup`) plus
  deterministic snippet expansion (`internal/snippets`) — it never rewrites the
  user's words with an LLM. Superwhisper-style "smart reformatting" needs an LLM
  and was explicitly declined because it adds latency. Don't reintroduce it
  without asking.
- **Local "AI-ish" features, no LLM/cloud:** custom **vocabulary** biases
  recognition via whisper's `initial_prompt` (built from `config.Vocabulary`,
  passed through the cgo shim); **snippets** are post-transcription whole-word
  text expansions (`config.Snippets`, applied in `app.finish`). Both are local
  and deterministic.
- **whisper.cpp with Metal + flash attention**, model `ggml-large-v3-turbo`.
  ~32× realtime on an M4 Pro. Flash attention is set in the cgo init
  (`internal/transcribe/cwhisper.c`), not a build flag.
- **Right-⌘ tap hotkey via a native `CGEventTap`** (`internal/hotkey`), not a
  key+modifier chord. It detects a *quick standalone tap* of the right Command
  key (keycode 54) and ignores it when used as a modifier in a shortcut, so the
  key keeps working normally. This is why we don't use `golang.design/x/hotkey`.
- **Paste = clipboard + synthesized Cmd+V, with focus restore** (`internal/paste`).
  At record start we remember the frontmost app and whether it can receive a
  paste (`HasPasteTarget`). On finish: paste into the frontmost app, *unless* it
  was the **Finder/desktop**, in which case show the result popup (and leave the
  text on the clipboard) instead of firing a Cmd+V that just beeps.
  - **Why bundle-id, not Accessibility:** we deliberately do **not** use the AX
    API to decide paste-vs-popup. Three AX signals were tried (focused element,
    settable value, focused window) and all get VS Code and the desktop
    *backwards*: Electron apps (VS Code) expose neither a focused element nor
    window, while the desktop (Finder) exposes both. So we paste into everything
    except when the frontmost app's bundle id is `com.apple.finder`. The
    frontmost app is captured on the **main thread** (`fyne.DoAndWait`) since
    even `NSWorkspace` is best used there. (Accessibility is still required, but
    only for the event tap and synthesizing Cmd+V — not for this decision.)
- **Zero-config / no settings UI.** The app is out-of-the-box: language defaults
  to `auto`, model to the app-support path. Advanced options (`model_path`,
  `language`, `vocabulary`, `snippets`) are read from `config.json` if present,
  but there is no settings window — don't re-add one without asking.
- **One cohesive `internal/app` package** owns the state machine and the Fyne UI
  (menu-bar item + result popup). UI mutations run on Fyne's main loop via
  `fyne.Do`. There is **no recording overlay window** — recording state is shown
  by the menu-bar icon (blinking red). The menu has a single **Quit** item.
- **Windows are forced light** with a **black accent** (the logo colour), not
  Fyne's default dark theme + blue. This is a deliberate custom theme
  (`internal/app/theme.go`, `silentTheme`) set in `New`; don't revert it to the
  system theme. The menu-bar tray icon is unaffected (it's a macOS template that
  adapts to the menu bar independently of the Fyne theme).

## Layout

```
cmd/silentrec        entrypoint (main)
internal/config      load/save settings (~/Library/Application Support/SilentRec/config.json)
internal/hotkey      right-⌘ tap detection (CGEventTap; chotkey.h/.m + hotkey.go)
internal/audio       mic capture via malgo → 16 kHz mono float32
internal/transcribe  whisper.cpp wrapper (cgo: cwhisper.h/.c + transcribe.go)
internal/cleanup     whitespace-only transcript tidy
internal/snippets    post-transcription text expansion (spoken phrase → replacement)
internal/paste       clipboard save/restore + Cmd+V + focus detection (cpaste.h/.m + paste.go)
internal/app         state machine + menu-bar item & result popup (app.go, ui.go, theme.go)
build/               Info.plist, entitlements, Makefile helpers, icon generator, signing script
build/icongen        renders icons from the S monogram
build/logo-source.svg  source of truth for the icon shape
third_party/whisper.cpp  vendored, built into static libs by `make whisper` (gitignored)
models/              downloaded model (gitignored)
```

## Build & run

```sh
make install        # everything: signing identity + whisper.cpp + model + build + install
make run            # same, but launches from dist/ instead of installing
make whisper        # (sub-step) clone + build whisper.cpp static libs (Metal)
make model          # (sub-step) download the model into ~/Library/Application Support/SilentRec/models
make icon           # force-regenerate all icons from build/logo-source.svg
```

`make install`/`run` are fully self-contained for a fresh clone: `sign` depends
on `signing-identity` (runs `build/setup-signing.sh`, idempotent — creates the
stable self-signed identity once), `build` depends on `$(WHISPER_LIB)` (clones +
builds whisper.cpp), and both depend on `model` (downloads the ~1.5 GB model,
skipped if present, into the app-support models dir the default config points
at). The model and `third_party/whisper.cpp` are gitignored, not committed.

Note: a bare `go build ./...` (without `make`) still needs the whisper static
libs to exist, since the cgo in `internal/transcribe` links them via
`${SRCDIR}`-relative paths — run `make whisper` once first.

## Critical gotchas

- **Code signing controls permission persistence.** macOS TCC keys permissions
  to the code-signing identity. Ad-hoc signing (`-`) changes the identity every
  build, so permissions reset every rebuild. We sign with a **stable self-signed
  identity** ("SilentRec Local Dev" in `silentrec-signing.keychain-db`, created
  by `build/setup-signing.sh`). The Makefile `sign` target uses it automatically.
  Don't switch to ad-hoc.
- **Two different permissions are involved:**
  - **Accessibility** — required for the right-⌘ event tap *and* for synthesizing
    Cmd+V *and* for focus/editable detection. This is the important one.
  - **Microphone** — for recording (prompted on first record).
  The app prompts for Accessibility on launch if missing.
- **Changing the bundle id resets all TCC permissions** (the user must re-grant).
  Avoid changing `com.kzkr.silentrec` unless necessary.
- **Launch via `open SilentRec.app`, not by running the binary directly.** Running
  the binary from a terminal makes macOS attribute permissions to the terminal,
  not the app, giving false "not trusted" results.
- **Menu-bar icon coloring:** Fyne renders a tray icon as an adaptive *template*
  (white on dark bar / dark on light) **only** if the resource is a
  `*theme.ThemedResource`; otherwise it uses literal pixel colors. The idle icon
  is wrapped in `theme.NewThemedResource` (adaptive); the recording icon is a
  plain `StaticResource` so its red shows as-is. See `internal/app/app.go`.
- **Embedded icons:** `internal/app/tray-idle.png` and `tray-rec.png` are
  `go:embed`-ed, so they must exist before `go build`. The Makefile regenerates
  them from `build/icongen` (which derives the shape from `logo-source.svg`).
- **whisper/ggml logging** is silenced via a no-op log callback in
  `cwhisper.c`; don't expect whisper progress on stderr.

## Conventions

- Keep it simple and minimal — Valentin values simplicity and may open-source this.
- Match the existing style: small single-responsibility packages, thin cgo shims
  with a clear C-symbol prefix (`silentrec_*`), comments explaining *why*.
- Run `go build ./... && go vet ./... && go test ./... && gofmt -l .` before
  considering a change done. The `-lobjc` duplicate-library linker warning is
  benign.

## Out of scope (deferred on purpose)

- Cross-platform (Linux/Windows) support.
- LLM-based cleanup / reformatting.
- Streaming transcription (we transcribe after stop).
- Overlay-over-fullscreen window (the menu-bar blink replaced the overlay).
