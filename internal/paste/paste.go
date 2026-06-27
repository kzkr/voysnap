// Package paste delivers transcribed text to the focused application.
//
// It restores focus to the app that was frontmost when recording started, puts
// the text on the clipboard, synthesizes Cmd+V, then restores the previous
// clipboard contents. All macOS interaction goes through a small Objective-C
// shim (cpaste.m).
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
	C.silentrec_remember_frontmost()
}

// AccessibilityTrusted reports whether SilentRec may synthesize keystrokes. When
// prompt is true and it is not trusted, macOS opens the permission dialog.
func AccessibilityTrusted(prompt bool) bool {
	return bool(C.silentrec_accessibility_trusted(C.bool(prompt)))
}

// SetClipboard replaces the clipboard text.
func SetClipboard(text string) {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	C.silentrec_clipboard_set(c)
}

// PasteText restores focus to the original app, puts text on the clipboard, and
// pastes it (Cmd+V). The transcript is intentionally left on the clipboard so it
// is recoverable if the paste lands nowhere (e.g. no text field is focused).
func PasteText(text string) {
	SetClipboard(text)
	C.silentrec_restore_frontmost()
	time.Sleep(120 * time.Millisecond) // let focus settle on the target app
	C.silentrec_paste()
}
