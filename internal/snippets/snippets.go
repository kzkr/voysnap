// Package snippets expands spoken trigger phrases into replacement text after
// transcription, e.g. saying "my email" yields "hello@kzkr.dev".
package snippets

import (
	"regexp"
	"strings"
)

// Expand replaces whole-word, case-insensitive occurrences of each trigger in
// text with its replacement. Returns text unchanged when m is empty.
func Expand(text string, m map[string]string) string {
	for trigger, replacement := range m {
		if strings.TrimSpace(trigger) == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(trigger) + `\b`)
		text = re.ReplaceAllLiteralString(text, replacement)
	}
	return text
}
