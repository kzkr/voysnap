#ifndef SILENTREC_CPASTE_H
#define SILENTREC_CPASTE_H

#include <stdbool.h>

// Replaces the clipboard contents with text.
void silentrec_clipboard_set(const char *text);

// Synthesizes a Cmd+V keystroke (requires Accessibility permission).
void silentrec_paste(void);

// Records the frontmost application so focus can be restored before pasting.
void silentrec_remember_frontmost(void);

// Reactivates the application remembered by silentrec_remember_frontmost.
void silentrec_restore_frontmost(void);

// Reports whether this process is trusted for Accessibility. When prompt is
// true, macOS shows the "open System Settings" dialog if it is not.
bool silentrec_accessibility_trusted(bool prompt);

#endif
