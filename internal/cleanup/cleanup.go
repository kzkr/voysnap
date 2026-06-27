// Package cleanup normalizes a raw transcript's whitespace. Whisper already
// punctuates and capitalizes, so this only trims the leading space Whisper adds
// and collapses any whitespace runs — it never changes your words.
package cleanup

import "strings"

// Clean trims surrounding whitespace and collapses internal whitespace runs to
// single spaces.
func Clean(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
