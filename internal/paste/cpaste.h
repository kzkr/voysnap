#ifndef VOYSNAP_CPASTE_H
#define VOYSNAP_CPASTE_H

#include <stdbool.h>

// Replaces the clipboard contents with text.
void voysnap_clipboard_set(const char *text);

// Synthesizes a Cmd+V keystroke (requires Accessibility permission).
void voysnap_paste(void);

// Records the frontmost application so focus can be restored before pasting.
void voysnap_remember_frontmost(void);

// Reactivates the application remembered by voysnap_remember_frontmost.
void voysnap_restore_frontmost(void);

// Reports whether the app frontmost at the last voysnap_remember_frontmost
// call was the Finder/desktop (i.e. there's nowhere to paste).
bool voysnap_frontmost_is_finder(void);

// Reports whether this process is trusted for Accessibility. When prompt is
// true, macOS shows the "open System Settings" dialog if it is not.
bool voysnap_accessibility_trusted(bool prompt);

#endif
