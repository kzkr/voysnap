#ifndef SILENTREC_CWHISPER_H
#define SILENTREC_CWHISPER_H

#include <stdbool.h>

// Thin C shim over the whisper.cpp API so the Go side never has to construct
// the large whisper_full_params struct directly.

// silentrec_init loads a ggml model. Returns an opaque context, or NULL on failure.
void *silentrec_init(const char *model_path, bool use_gpu);

// silentrec_transcribe runs whisper over mono 16 kHz float samples and writes a
// newly-allocated, NUL-terminated transcript to *out_text (free with
// silentrec_free_text). language may be "en" or NULL for auto-detect. prompt is
// an optional vocabulary hint (NULL/empty to skip). Returns 0 on success,
// non-zero on failure.
int silentrec_transcribe(void *ctx, const float *samples, int n_samples,
                     const char *language, const char *prompt, int n_threads,
                     char **out_text);

// silentrec_free releases a context returned by silentrec_init.
void silentrec_free(void *ctx);

// silentrec_free_text frees a transcript returned by silentrec_transcribe.
void silentrec_free_text(char *text);

#endif
