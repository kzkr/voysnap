#include "cwhisper.h"
#include "whisper.h"
#include "ggml.h"

#include <stdlib.h>
#include <string.h>

// Swallow whisper/ggml's verbose progress logging; the app surfaces state
// through its own UI instead.
static void silentrec_silent_log(enum ggml_log_level level, const char *text,
                             void *user_data) {
  (void)level;
  (void)text;
  (void)user_data;
}

void *silentrec_init(const char *model_path, bool use_gpu) {
  whisper_log_set(silentrec_silent_log, NULL);
  ggml_log_set(silentrec_silent_log, NULL);

  struct whisper_context_params cparams = whisper_context_default_params();
  cparams.use_gpu = use_gpu;
  cparams.flash_attn = use_gpu; // big speedup on Metal, esp. for long audio
  return (void *)whisper_init_from_file_with_params(model_path, cparams);
}

int silentrec_transcribe(void *vctx, const float *samples, int n_samples,
                     const char *language, const char *prompt, int n_threads,
                     char **out_text) {
  struct whisper_context *ctx = (struct whisper_context *)vctx;

  struct whisper_full_params params =
      whisper_full_default_params(WHISPER_SAMPLING_GREEDY);
  params.print_realtime = false;
  params.print_progress = false;
  params.print_timestamps = false;
  params.print_special = false;
  params.translate = false;
  params.language = language; // "en", or NULL to auto-detect
  params.n_threads = n_threads;
  params.no_context = true;     // each dictation is independent
  params.suppress_blank = true;
  if (prompt != NULL && prompt[0] != '\0') {
    params.initial_prompt = prompt; // bias toward custom vocabulary
  }

  if (whisper_full(ctx, params, samples, n_samples) != 0) {
    return -1;
  }

  // Concatenate all segment texts into one growable buffer.
  int n = whisper_full_n_segments(ctx);
  size_t cap = 256, len = 0;
  char *buf = (char *)malloc(cap);
  if (buf == NULL) {
    return -2;
  }
  buf[0] = '\0';

  for (int i = 0; i < n; i++) {
    const char *seg = whisper_full_get_segment_text(ctx, i);
    size_t sl = strlen(seg);
    if (len + sl + 1 > cap) {
      while (len + sl + 1 > cap) {
        cap *= 2;
      }
      char *grown = (char *)realloc(buf, cap);
      if (grown == NULL) {
        free(buf);
        return -2;
      }
      buf = grown;
    }
    memcpy(buf + len, seg, sl);
    len += sl;
    buf[len] = '\0';
  }

  *out_text = buf;
  return 0;
}

void silentrec_free(void *vctx) {
  whisper_free((struct whisper_context *)vctx);
}

void silentrec_free_text(char *text) { free(text); }
