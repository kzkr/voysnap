#ifndef SILENTREC_CPASTE_H
#define SILENTREC_CPASTE_H

#include <stdbool.h>

// Returns the current clipboard text (malloc'd, free with silentrec_str_free), or
// NULL if the clipboard holds no string.
char *silentrec_clipboard_get(void);

// Replaces the clipboard contents with text.
void silentrec_clipboard_set(const char *text);

// Synthesizes a Cmd+V keystroke (requires Accessibility permission).
void silentrec_paste(void);

// Records the frontmost application (so focus can be restored before pasting)
// and whether its focused UI element is an editable text field.
void silentrec_remember_frontmost(void);

// Reports whether the element focused at the last silentrec_remember_frontmost call
// was an editable text field. Requires Accessibility permission.
bool silentrec_focused_was_editable(void);

// Reactivates the application remembered by silentrec_remember_frontmost.
void silentrec_restore_frontmost(void);

// Reports whether this process is trusted for Accessibility. When prompt is
// true, macOS shows the "open System Settings" dialog if it is not.
bool silentrec_accessibility_trusted(bool prompt);

void silentrec_str_free(char *s);

#endif
