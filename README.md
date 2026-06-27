# SilentRec

A free, fully-local, offline voice-to-text dictation app for macOS. Tap the
**right ⌘ key** (to the right of the space bar) to start recording, tap it again
to stop — the audio is transcribed on-device with [whisper.cpp](https://github.com/ggerganov/whisper.cpp)
(Metal-accelerated) and pasted into whatever text field you're in. If nothing is
focused, the transcript pops up in a small window. Either way it's also left on
the clipboard.

Speak **any language** — it auto-detects (English, French, Spanish, …) and writes
it back in that language. No cloud, no accounts, no word limits, no subscription.

## Requirements

- macOS 13+ on Apple Silicon
- [Homebrew](https://brew.sh), Go 1.24+, `cmake` (`brew install cmake`), and the
  Xcode command-line tools (`xcode-select --install`)

## Install

```sh
git clone git@github.com:kzkr/silentrec.git
cd silentrec
make install
```

That's it. `make install` does everything: creates a stable signing identity,
builds whisper.cpp (Metal), downloads the model (~1.5 GB), builds the app, and
installs **SilentRec.app** to `/Applications`. The first run is slow (it compiles
whisper.cpp and downloads the model into `~/Library/Application Support/SilentRec/`);
later builds skip both.

Use `make run` instead to launch from `dist/` without installing.

Launch SilentRec like any other app (Launchpad/Spotlight) and keep it in the
Dock if you like. A microphone icon appears in the menu bar.

## Permissions (granted once)

On first launch macOS asks for two permissions:

- **Microphone** — to record your voice.
- **Accessibility** — to detect the right-⌘ tap and to paste via Cmd+V. Enable
  **SilentRec** under *System Settings → Privacy & Security → Accessibility*,
  then relaunch.

Because the app is signed with a stable identity, these grants persist across
rebuilds.

## Usage

1. Put your cursor in a text field (any app).
2. Tap the **right ⌘** key — the menu-bar icon blinks red.
3. Speak.
4. Tap **right ⌘** again — SilentRec transcribes and pastes the text.

The right ⌘ key still works normally as a modifier in shortcuts; only a quick
standalone tap triggers dictation. If nothing is focused (e.g. the desktop), the
transcript appears in a popup instead of being pasted.

The text is pasted exactly as Whisper produces it (it already punctuates and
capitalizes) — the only processing is whitespace trimming and your snippet
expansions. No LLM ever rewrites your words.

## Configuration

SilentRec is zero-config — it works out of the box (auto-detect language, paste
where your cursor is). There is no settings window. Power users can optionally
edit `~/Library/Application Support/SilentRec/config.json`:

- `language` — `"auto"` (default) or a code like `"en"` / `"fr"`.
- `model_path` — path to a different ggml model.
- `vocabulary` — array of names/jargon to recognize better (whisper prompt bias).
- `snippets` — object of `"spoken phrase": "replacement"` text expansions.

## Project layout

```
cmd/silentrec        entrypoint
internal/config      load settings (zero-config defaults; optional config.json)
internal/hotkey      right-⌘ tap detection (CGEventTap)
internal/audio       microphone capture (malgo) → 16 kHz mono
internal/transcribe  whisper.cpp wrapper (cgo, static libs)
internal/cleanup     whitespace trimming of the transcript
internal/snippets    spoken-phrase → replacement text expansion
internal/paste       clipboard + synthesized Cmd+V; popup fallback on the desktop
internal/app         state machine + menu-bar item & result popup
build/               Info.plist, signing script, icon generator, logo source
third_party          vendored whisper.cpp (built by make; gitignored)
```

## License

Personal project by [kzkr](https://github.com/kzkr) — South Computer Company.
