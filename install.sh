#!/usr/bin/env bash
#
# VoySnap installer — downloads the prebuilt app and the speech model, then
# launches it. No Xcode, no Go, no build step.
#
#   curl -fsSL https://raw.githubusercontent.com/kzkr/voysnap/main/install.sh | bash
#
# Installing via curl (rather than a browser) means the download is not
# quarantined, so macOS Gatekeeper does not block the app.

set -euo pipefail

APP_NAME="VoySnap"
REPO="kzkr/voysnap"
ASSET="VoySnap.zip"
RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

APP_DST="/Applications/${APP_NAME}.app"
EXECUTABLE="voysnap"

SUPPORT="${HOME}/Library/Application Support/${APP_NAME}"
MODELS_DIR="${SUPPORT}/models"
MODEL_NAME="ggml-large-v3-turbo.bin"
MODEL_PATH="${MODELS_DIR}/${MODEL_NAME}"
MODEL_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main/${MODEL_NAME}"

# --- platform check ----------------------------------------------------------
if [ "$(uname -s)" != "Darwin" ] || [ "$(uname -m)" != "arm64" ]; then
  echo "VoySnap requires macOS on Apple Silicon (arm64)." >&2
  exit 1
fi

# --- download + install the app ----------------------------------------------
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "==> Downloading ${APP_NAME}…"
curl -fSL --progress-bar -o "${TMP}/${ASSET}" "${RELEASE_URL}"

echo "==> Installing to ${APP_DST}…"
killall "${EXECUTABLE}" 2>/dev/null || true   # quit a running instance, if any
rm -rf "${APP_DST}"
ditto -x -k "${TMP}/${ASSET}" /Applications
# curl downloads aren't quarantined, but strip it defensively anyway.
xattr -dr com.apple.quarantine "${APP_DST}" 2>/dev/null || true

# --- download the model (once) -----------------------------------------------
if [ -f "${MODEL_PATH}" ]; then
  echo "==> Speech model already present."
else
  echo "==> Downloading speech model (~1.5 GB, one time)…"
  mkdir -p "${MODELS_DIR}"
  curl -fSL --progress-bar -o "${MODEL_PATH}" "${MODEL_URL}"
fi

# --- launch ------------------------------------------------------------------
echo "==> Launching ${APP_NAME}…"
open "${APP_DST}"

cat <<EOF

${APP_NAME} is installed. 🎙️

  • Look for the icon in your menu bar.
  • Grant Microphone and Accessibility when macOS asks — Accessibility powers
    the right-⌘ hotkey and pasting.
  • Put your cursor anywhere, tap the right ⌘ key, speak, and tap it again.

EOF
