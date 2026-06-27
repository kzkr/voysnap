// Package transcribe wraps whisper.cpp to turn 16 kHz mono audio into text.
//
// The vendored whisper.cpp static libraries (built via `make whisper`) are
// linked directly through cgo using ${SRCDIR}-relative paths, so `go build`
// needs no environment variables.
package transcribe

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/whisper.cpp/include -I${SRCDIR}/../../third_party/whisper.cpp/ggml/include
#cgo LDFLAGS: ${SRCDIR}/../../third_party/whisper.cpp/build/src/libwhisper.a ${SRCDIR}/../../third_party/whisper.cpp/build/ggml/src/ggml-metal/libggml-metal.a ${SRCDIR}/../../third_party/whisper.cpp/build/ggml/src/ggml-blas/libggml-blas.a ${SRCDIR}/../../third_party/whisper.cpp/build/ggml/src/libggml-cpu.a ${SRCDIR}/../../third_party/whisper.cpp/build/ggml/src/libggml.a ${SRCDIR}/../../third_party/whisper.cpp/build/ggml/src/libggml-base.a -lstdc++ -framework Foundation -framework Metal -framework MetalKit -framework Accelerate -framework CoreFoundation
#include <stdlib.h>
#include "cwhisper.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Transcriber holds a loaded model. It is safe for one transcription at a time;
// callers must not invoke Transcribe concurrently on the same instance.
type Transcriber struct {
	mu  sync.Mutex
	ctx unsafe.Pointer
}

// Load reads a ggml model file and initializes whisper with the GPU (Metal)
// backend enabled.
func Load(modelPath string) (*Transcriber, error) {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	ctx := C.silentrec_init(cPath, C.bool(true))
	if ctx == nil {
		return nil, fmt.Errorf("transcribe: failed to load model %q", modelPath)
	}
	return &Transcriber{ctx: ctx}, nil
}

// Transcribe converts mono 16 kHz float32 samples (range -1..1) into text.
// language is a hint such as "en"; pass "" or "auto" to auto-detect. prompt is
// an optional vocabulary hint to bias recognition ("" to skip).
func (t *Transcriber) Transcribe(samples []float32, language, prompt string) (string, error) {
	if len(samples) == 0 {
		return "", nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx == nil {
		return "", errors.New("transcribe: model is closed")
	}

	var cLang *C.char
	if language != "" && language != "auto" {
		cLang = C.CString(language)
		defer C.free(unsafe.Pointer(cLang))
	}

	var cPrompt *C.char
	if prompt != "" {
		cPrompt = C.CString(prompt)
		defer C.free(unsafe.Pointer(cPrompt))
	}

	threads := runtime.NumCPU()
	if threads > 8 {
		threads = 8 // diminishing returns; the heavy lifting is on the GPU
	}

	var out *C.char
	rc := C.silentrec_transcribe(
		t.ctx,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(len(samples)),
		cLang,
		cPrompt,
		C.int(threads),
		&out,
	)
	if rc != 0 {
		return "", fmt.Errorf("transcribe: whisper_full failed (code %d)", int(rc))
	}
	defer C.silentrec_free_text(out)

	return C.GoString(out), nil
}

// Close releases the underlying model.
func (t *Transcriber) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx != nil {
		C.silentrec_free(t.ctx)
		t.ctx = nil
	}
}
