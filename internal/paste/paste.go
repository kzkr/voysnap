// Package paste delivers transcribed text to the frontmost application.
//
// It restores focus to the app that was frontmost when recording started, puts
// the text on the clipboard, and synthesizes Cmd+V. The text is left on the
// clipboard as a fallback. All macOS interaction goes through a small
// Objective-C shim (cpaste.m).
package paste

/*
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices -framework CoreGraphics
#include <stdlib.h>
#include "cpaste.h"
*/
import "C"

import (
	"time"
	"unsafe"
)

// RememberTarget records the frontmost app so focus can be restored before
// pasting. Call this the moment recording starts.
func RememberTarget() {
	C.voysnap_remember_frontmost()
}

// HasPasteTarget reports whether the frontmost app (at the last RememberTarget)
// can receive a paste — true for normal apps, false for the Finder/desktop.
// When false, show the transcript in a popup instead of pasting.
func HasPasteTarget() bool {
	return !bool(C.voysnap_frontmost_is_finder())
}

// AccessibilityTrusted reports whether Voysnap may synthesize keystrokes. When
// prompt is true and it is not trusted, macOS opens the permission dialog.
func AccessibilityTrusted(prompt bool) bool {
	return bool(C.voysnap_accessibility_trusted(C.bool(prompt)))
}

// SetClipboard replaces the clipboard text.
func SetClipboard(text string) {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	C.voysnap_clipboard_set(c)
}

// PasteText restores focus to the original app, puts text on the clipboard, and
// pastes it (Cmd+V). The transcript is intentionally left on the clipboard so it
// is recoverable if the paste lands nowhere (e.g. no text field is focused).
func PasteText(text string) {
	SetClipboard(text)
	C.voysnap_restore_frontmost()
	time.Sleep(120 * time.Millisecond) // let focus settle on the target app
	C.voysnap_paste()
}
