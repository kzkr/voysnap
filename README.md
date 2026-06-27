# Voysnap

**Free, fully-local, offline voice-to-text dictation for macOS.**

Tap the **right ⌘ key**, speak, tap it again — your words are transcribed
on-device and pasted into whatever app you're in. No cloud, no account, no word
limits, no subscription.

## Features

- 🎙️ **Dictate anywhere** — pastes straight into the focused app (or pops up the
  text when you're on the desktop).
- 🌍 **Multilingual** — auto-detects the language and writes it back in that
  language (English, French, Spanish, …).
- ⚡ **Fast** — [whisper.cpp](https://github.com/ggerganov/whisper.cpp) with Metal
  on Apple Silicon, ~30× faster than real time.
- 🔒 **100% private** — runs entirely offline; audio never leaves your Mac.
- 🪶 **Zero-config** — a tiny menu-bar app that just works.

## Install

Requires macOS 13+ (Apple Silicon), [Homebrew](https://brew.sh), Go 1.24+,
`cmake`, and the Xcode command-line tools (`xcode-select --install`).

```sh
git clone git@github.com:kzkr/voysnap.git
cd voysnap
make install
```

`make install` does everything — builds whisper.cpp, downloads the model
(~1.5 GB), and installs **Voysnap.app** to `/Applications`. The first build
takes a few minutes; after that it's instant.

On first launch, grant **Microphone** and **Accessibility** when asked (the
latter powers the hotkey and pasting). That's it.

## Usage

1. Put your cursor where you want to type.
2. Tap the **right ⌘** key — the menu-bar icon turns red.
3. Speak.
4. Tap **right ⌘** again — your text appears.

A quick standalone tap triggers dictation; the right ⌘ key still works normally
as a modifier in shortcuts. Text is pasted exactly as transcribed (whisper
punctuates and capitalizes) and also left on the clipboard.

## Configuration

Voysnap needs no setup. Power users can optionally edit
`~/Library/Application Support/Voysnap/config.json`:

| Key | Description |
| --- | --- |
| `language` | `"auto"` (default), or a code like `"en"` / `"fr"` |
| `model_path` | path to a different ggml model |
| `vocabulary` | array of names/jargon to recognize better |
| `snippets` | `{ "spoken phrase": "replacement" }` text expansions |

## How it works

A small Go + cgo menu-bar app. A native `CGEventTap` detects the right-⌘ tap,
`malgo` captures the mic at 16 kHz, whisper.cpp (Metal) transcribes it, and the
result is pasted via a synthesized Cmd+V. See [CLAUDE.md](CLAUDE.md) for the
architecture and design notes.

## License

[MIT](LICENSE) © [kzkr](https://github.com/kzkr) — a South Computer Company project.
