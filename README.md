# VoySnap

> **Press. Speak. Done.**

Free, open-source, fully local voice dictation for macOS.

Press the **right ⌘ key**, speak naturally, press it again — your words are
transcribed on-device and pasted straight into whatever app you're using.

**No cloud. No account. No subscriptions. No word limits.**

![Platform: macOS Apple Silicon](https://img.shields.io/badge/platform-macOS%20·%20Apple%20Silicon-black)
![License: MIT](https://img.shields.io/badge/license-MIT-black)

## Features

- 🎙️ **Dictate anywhere** — pastes into the focused text field of any app, or shows a popup when you're on the desktop.
- ⚡ **Blazing fast** — [`whisper.cpp`](https://github.com/ggerganov/whisper.cpp) with Metal on Apple Silicon: over **30× faster than real time**.
- 🔒 **100% private** — runs entirely on your Mac. Your voice never leaves your device.
- 🌍 **Multilingual** — auto-detects the language you speak and writes it back in that language.
- 🪶 **Minimal** — a tiny menu-bar app. Zero onboarding, zero accounts, zero distractions.

## Installation

**Requirements:** macOS 13+ · Apple Silicon · [Homebrew](https://brew.sh) · Go 1.24+ · CMake · Xcode Command Line Tools

```bash
xcode-select --install        # if you don't have the Xcode tools yet

git clone git@github.com:kzkr/voysnap.git
cd voysnap
make install
```

`make install` does everything: builds `whisper.cpp`, downloads the model
(~1.5 GB), and installs **VoySnap.app** into `/Applications`. The first build
takes a few minutes; after that it's nearly instant.

On first launch, macOS asks for two permissions:

- 🎤 **Microphone** — to hear you.
- ♿ **Accessibility** — for the global hotkey and automatic pasting.

That's it.

## Usage

1. Put your cursor where you want to type.
2. Tap the **right ⌘** key — the menu-bar icon turns red.
3. Speak.
4. Tap **right ⌘** again — your text appears.

A quick standalone tap toggles dictation; using right ⌘ as a modifier in
keyboard shortcuts still works normally. Text is pasted exactly as transcribed —
whisper handles punctuation and capitalization — and also left on your clipboard.

## Configuration

VoySnap works out of the box. To customize it, edit:

```text
~/Library/Application Support/VoySnap/config.json
```

| Key          | Description                                                      |
| ------------ | ---------------------------------------------------------------- |
| `language`   | `"auto"` (default), or a code such as `"en"` / `"fr"`            |
| `model_path` | path to a different whisper model                                |
| `vocabulary` | custom words, names, or jargon to recognize better               |
| `snippets`   | `{ "spoken phrase": "replacement" }` text expansions             |

## How it works

VoySnap is a lightweight Go + `cgo` menu-bar app:

- a native `CGEventTap` detects the **right ⌘** tap,
- `malgo` captures the mic at 16 kHz,
- `whisper.cpp` transcribes locally with Metal,
- the result is pasted via a synthesized `⌘V`.

See [CLAUDE.md](CLAUDE.md) for architecture and design notes.

## Why VoySnap?

Built-in macOS dictation leans on Apple's services, and most AI dictation tools
want an account, a subscription, or your audio in their cloud.

VoySnap is different:

- ✅ Runs entirely on your Mac
- ✅ Free and open source
- ✅ Works in every application
- ✅ No accounts, subscriptions, or usage limits

Just **press, speak, done.**

## License

[MIT](LICENSE) © [kzkr](https://github.com/kzkr) — a South Computer Company project.
