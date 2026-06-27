#ifndef VOYSNAP_CWHISPER_H
#define VOYSNAP_CWHISPER_H

#include <stdbool.h>

// Thin C shim over the whisper.cpp API so the Go side never has to construct
// the large whisper_full_params struct directly.

// voysnap_init loads a ggml model. Returns an opaque context, or NULL on failure.
void *voysnap_init(const char *model_path, bool use_gpu);

// voysnap_transcribe runs whisper over mono 16 kHz float samples and writes a
// newly-allocated, NUL-terminated transcript to *out_text (free with
// voysnap_free_text). language may be "en" or NULL for auto-detect. prompt is
// an optional vocabulary hint (NULL/empty to skip). Returns 0 on success,
// non-zero on failure.
int voysnap_transcribe(void *ctx, const float *samples, int n_samples,
                     const char *language, const char *prompt, int n_threads,
                     char **out_text);

// voysnap_free releases a context returned by voysnap_init.
void voysnap_free(void *ctx);

// voysnap_free_text frees a transcript returned by voysnap_transcribe.
void voysnap_free_text(char *text);

#endif
