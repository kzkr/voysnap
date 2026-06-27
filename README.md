# SilentRec

A free, fully-local, offline voice-to-text dictation app for macOS. Tap the
**right ⌘ key** (to the right of the space bar) to start recording, tap it again
to stop — the audio is transcribed on-device with [whisper.cpp](https://github.com/ggerganov/whisper.cpp)
(Metal-accelerated), lightly cleaned up, and pasted into whatever text field you
were focused on. If no field is focused, the text appears in a popup (and is
copied to the clipboard).

No cloud, no accounts, no word limits, no subscription.

## Requirements

- macOS 13+ on Apple Silicon
- Go 1.24+, `cmake`, Xcode command-line tools (`xcode-select --install`)

## Build & run

```sh
./build/setup-signing.sh  # one-time: stable signing identity (keeps permissions across rebuilds)
make install              # builds whisper.cpp, pulls the model (~1.5GB), builds & installs the app
# or: make run            # same, but launches from dist/ without installing
```

The first build is slow: it clones and compiles whisper.cpp and downloads the
~1.5 GB model into `~/Library/Application Support/SilentRec/models/`. Subsequent
builds skip both. Point at a different model any time in **Settings**.

Once installed, launch SilentRec like any other app and keep it in the Dock. A
microphone icon appears in the menu bar.

## Permissions (granted once)

On first use macOS will ask for two permissions:

- **Microphone** — to record your voice (prompted the first time you record).
- **Accessibility** — to detect the right-⌘ tap and to paste via Cmd+V.
  Grant it under *System Settings → Privacy & Security → Accessibility*, then
  relaunch SilentRec.

## Usage

1. Focus a text field (e.g. in VS Code).
2. Tap the **right ⌘** key — the menu-bar icon blinks red.
3. Speak.
4. Tap **right ⌘** again — SilentRec transcribes and pastes the text.

The right ⌘ key still works normally as a modifier in keyboard shortcuts; only a
quick standalone tap triggers dictation.

## Settings

Open **Settings…** from the menu-bar icon to configure:

- **Model** — path to the ggml model (with a file picker).
- **Language** — *Auto-detect* (default) or a specific language. The model is
  multilingual, so auto-detect transcribes whatever you speak.
- **Custom words** — names/jargon/acronyms to recognize more accurately
  (biases whisper via its initial prompt).
- **Snippets** — spoken phrase → replacement text, e.g. `my email = hello@kzkr.dev`.

The transcript is always auto-pasted into the focused field (or shown in a popup
if nothing editable is focused), as Whisper produces it (it already punctuates
and capitalizes) — the only processing is whitespace trimming and your snippet
expansions. No LLM ever rewrites your words.

## Project layout

```
cmd/silentrec        entrypoint
internal/config      load/save settings
internal/hotkey      right-⌘ tap detection (CGEventTap)
internal/audio       microphone capture (malgo) → 16 kHz mono
internal/transcribe  whisper.cpp wrapper (cgo, static libs)
internal/cleanup     whitespace trimming of the transcript
internal/snippets    spoken-phrase → replacement text expansion
internal/paste       clipboard save/restore + synthesized Cmd+V
internal/app         state machine + menu-bar/overlay/popup/settings UI
third_party          vendored whisper.cpp (built by `make whisper`)
```
