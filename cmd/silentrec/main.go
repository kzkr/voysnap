// Command silentrec is a local, offline voice-to-text dictation menu-bar app for macOS.
//
// Flow: tap the right ⌘ key to start recording, tap it again to stop. The audio
// is transcribed locally with whisper.cpp, lightly cleaned up, and pasted into
// the focused text field (or shown in a popup when nothing is focused).
package main

import (
	"log"

	"github.com/kzkr/silentrec/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	a.Run()
}
