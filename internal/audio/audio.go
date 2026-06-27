// Package audio captures microphone input as 16 kHz mono float32 samples,
// the format whisper.cpp expects. miniaudio (via malgo) resamples from the
// device's native rate for us.
package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
)

const (
	sampleRate = 16000
	channels   = 1
)

// Recorder owns a miniaudio context and a capture device. Use Start to begin
// buffering samples and Stop to retrieve them. Not safe for concurrent use.
type Recorder struct {
	ctx    *malgo.AllocatedContext
	device *malgo.Device

	mu        sync.Mutex
	buf       []float32
	recording bool
}

// NewRecorder initializes the audio backend. Call Close when done.
func NewRecorder() (*Recorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("audio: init context: %w", err)
	}

	r := &Recorder{ctx: ctx}

	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = channels
	cfg.SampleRate = sampleRate

	device, err := malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{
		Data: r.onFrames,
	})
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return nil, fmt.Errorf("audio: init device: %w", err)
	}
	r.device = device
	return r, nil
}

// onFrames runs on the audio thread; it appends captured samples while recording.
func (r *Recorder) onFrames(_, input []byte, frameCount uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return
	}
	n := int(frameCount) * channels
	for i := 0; i < n && (i+1)*4 <= len(input); i++ {
		bits := binary.LittleEndian.Uint32(input[i*4 : i*4+4])
		r.buf = append(r.buf, math.Float32frombits(bits))
	}
}

// Start clears the buffer and begins capturing. Triggers the macOS microphone
// permission prompt on first use.
func (r *Recorder) Start() error {
	r.mu.Lock()
	r.buf = r.buf[:0]
	r.recording = true
	r.mu.Unlock()

	if err := r.device.Start(); err != nil {
		return fmt.Errorf("audio: start device: %w", err)
	}
	return nil
}

// Stop halts capture and returns the recorded mono 16 kHz samples.
func (r *Recorder) Stop() ([]float32, error) {
	r.mu.Lock()
	r.recording = false
	r.mu.Unlock()

	if err := r.device.Stop(); err != nil {
		return nil, fmt.Errorf("audio: stop device: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]float32, len(r.buf))
	copy(out, r.buf)
	return out, nil
}

// Close releases the device and context.
func (r *Recorder) Close() {
	if r.device != nil {
		r.device.Uninit()
	}
	if r.ctx != nil {
		_ = r.ctx.Uninit()
		r.ctx.Free()
	}
}
