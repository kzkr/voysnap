// Package config loads and persists user settings for Voysnap.
//
// Settings live in ~/Library/Application Support/Voysnap/config.json. Missing or
// unset fields fall back to sensible defaults, so a fresh install works with no
// configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	appName          = "Voysnap"
	defaultModelName = "ggml-large-v3-turbo.bin"
)

// Config holds all user-tunable settings.
type Config struct {
	// Hotkey is the key whose tap toggles recording on/off.
	Hotkey Hotkey `json:"hotkey"`

	// ModelPath is the absolute path to the whisper.cpp ggml model file.
	ModelPath string `json:"model_path"`

	// Language is the spoken language: "auto" to detect per utterance, or a code
	// like "en"/"fr".
	Language string `json:"language"`

	// Vocabulary biases recognition toward these words/phrases (names, jargon,
	// acronyms) via whisper's initial prompt.
	Vocabulary []string `json:"vocabulary"`

	// Snippets expand a spoken phrase into replacement text after transcription
	// (e.g. "my email" -> "hello@kzkr.dev").
	Snippets map[string]string `json:"snippets"`
}

// Hotkey names the key whose tap toggles recording, e.g. {Key: "right_command"}.
type Hotkey struct {
	Key string `json:"key"`
}

// Default returns the built-in configuration used when nothing is saved yet.
func Default() Config {
	return Config{
		Hotkey: Hotkey{
			// Tap the right Command key (to the right of the space bar) to
			// start/stop dictation.
			Key: "right_command",
		},
		ModelPath: DefaultModelPath(),
		Language:  "auto",
	}
}

// Dir is the application support directory for Voysnap, created on demand.
func Dir() (string, error) {
	base, err := os.UserConfigDir() // ~/Library/Application Support on macOS
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// DefaultModelPath returns the default location for the speech model.
func DefaultModelPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return defaultModelName
	}
	return filepath.Join(base, appName, "models", defaultModelName)
}

func path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the saved config, filling any unset fields with defaults. If no
// config file exists yet it returns the defaults (without writing a file).
func Load() (Config, error) {
	cfg := Default()
	p, err := path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", p, err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

// Save writes the config to disk as indented JSON.
func (c Config) Save() error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// applyDefaults backfills empty fields after unmarshalling a partial file.
func (c *Config) applyDefaults() {
	def := Default()
	if c.Hotkey.Key == "" {
		c.Hotkey = def.Hotkey
	}
	if c.ModelPath == "" {
		c.ModelPath = def.ModelPath
	}
	if c.Language == "" {
		c.Language = def.Language
	}
}
