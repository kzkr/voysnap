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

// RememberTarget records the frontmost app and whether its focused element is
// an editable text field. Call this the moment recording starts, before any
// SilentRec window can steal focus.
func RememberTarget() {
	C.silentrec_remember_frontmost()
}

// FocusedWasEditable reports whether an editable text field was focused at the
// last RememberTarget call.
func FocusedWasEditable() bool {
	return bool(C.silentrec_focused_was_editable())
}

// AccessibilityTrusted reports whether SilentRec may synthesize keystrokes. When
// prompt is true and it is not trusted, macOS opens the permission dialog.
func AccessibilityTrusted(prompt bool) bool {
	return bool(C.silentrec_accessibility_trusted(C.bool(prompt)))
}

// GetClipboard returns the current clipboard text.
func GetClipboard() string {
	c := C.silentrec_clipboard_get()
	if c == nil {
		return ""
	}
	defer C.silentrec_str_free(c)
	return C.GoString(c)
}

// SetClipboard replaces the clipboard text.
func SetClipboard(text string) {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	C.silentrec_clipboard_set(c)
}

// PasteText restores focus to the original app and pastes text, then restores
// the clipboard to whatever it held before.
func PasteText(text string) {
	prev := GetClipboard()

	SetClipboard(text)
	C.silentrec_restore_frontmost()
	time.Sleep(120 * time.Millisecond) // let focus settle on the target app

	C.silentrec_paste()
	time.Sleep(150 * time.Millisecond) // let the target read the clipboard

	SetClipboard(prev)
}
