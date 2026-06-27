<p align="center">
  <img height="100" src="https://github.com/user-attachments/assets/0c9c9b26-4333-43c2-abf3-1d9ba516778d" alt="VoySnap — press the right Command key, speak, and your words are typed for you."/>
</p>

<h1 align="center">VoySnap</h1>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-black?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/platform-macOS%20·%20Apple%20Silicon-black?style=flat-square" alt="Platform">
  <img src="https://img.shields.io/badge/100%25-local-black?style=flat-square" alt="100% local">
</p>

Your voice belongs on your Mac.

Press the **right ⌘ key**, speak naturally, press it again, and your words are transcribed on-device and pasted into whatever app you're using.

No cloud. No account. No subscriptions. No word limits.

## How it works

**1. Press to record** — a native `CGEventTap` catches a quick tap of the right ⌘ key. The menu-bar icon turns red.

**2. Speak** — `malgo` captures your mic at 16 kHz, fully on-device.

**3. Press again** — `whisper.cpp` transcribes locally with Metal acceleration, over **30× faster than real time**.

**4. Done** — your text is pasted into the active app and copied to your clipboard.

Nothing leaves your Mac. The right ⌘ key still works normally as a modifier in keyboard shortcuts — only a quick standalone tap triggers dictation.

## Why VoySnap

🔒 **Private by design** — every transcription happens on your Mac. Your voice never leaves your device.

💸 **Free forever** — open source, no subscriptions, no API keys, no limits.

⚡ **Blazing fast** — Metal-accelerated `whisper.cpp` on Apple Silicon transcribes faster than you can re-read it.

🌍 **Multilingual** — auto-detects the language you speak and writes it back in that language.

🪶 **Invisible** — a tiny menu-bar app. Zero onboarding, zero accounts, zero distractions.

## Install

```bash
git clone git@github.com:kzkr/voysnap.git
cd voysnap
make install
```

`make install` does everything: builds `whisper.cpp`, downloads the model (~1.5 GB), and installs **VoySnap.app** into `/Applications`. The first build takes a few minutes; after that it's nearly instant.

On first launch, macOS asks for two permissions:

- 🎤 **Microphone** — to hear you.
- ♿ **Accessibility** — for the global hotkey and automatic pasting.

That's it.

<details>
<summary>Requirements</summary>
<br>

- macOS 13+ · Apple Silicon
- [Homebrew](https://brew.sh)
- Go 1.24+
- CMake
- Xcode Command Line Tools (`xcode-select --install`)

</details>

## Usage

1. Put your cursor where you want to type.
2. Tap the **right ⌘** key — the menu-bar icon turns red.
3. Speak.
4. Tap **right ⌘** again — your text appears.

Text is pasted exactly as Whisper transcribes it, including punctuation and capitalization. When nothing editable is focused — e.g. on the desktop — VoySnap shows the result in a popup and leaves it on your clipboard instead.

## Configuration

VoySnap works out of the box. To customize it, edit `~/Library/Application Support/VoySnap/config.json`:

| Key          | Description                                              |
| ------------ | -------------------------------------------------------- |
| `language`   | `"auto"` (default), or a code such as `"en"` / `"fr"`    |
| `model_path` | path to a different whisper model                        |
| `vocabulary` | custom words, names, or jargon to recognize better       |
| `snippets`   | `{ "spoken phrase": "replacement" }` text expansions     |

All transcription runs locally via [`whisper.cpp`](https://github.com/ggerganov/whisper.cpp). See [CLAUDE.md](CLAUDE.md) for architecture and design notes.

## License

[MIT](LICENSE)
