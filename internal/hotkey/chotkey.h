#ifndef SILENTREC_CHOTKEY_H
#define SILENTREC_CHOTKEY_H

// Creates a listen-only event tap watching the given modifier keycode (e.g. 54
// for the right Command key). Returns 0 on success, -1 if the process is not
// trusted for Accessibility (so the tap could not be created).
int silentrec_hotkey_create(int keycode);

// Runs the tap's run loop. Blocks until silentrec_hotkey_stop is called; must be
// invoked on the same (locked) thread that called silentrec_hotkey_create.
void silentrec_hotkey_run(void);

// Stops the run loop started by silentrec_hotkey_run.
void silentrec_hotkey_stop(void);

#endif
