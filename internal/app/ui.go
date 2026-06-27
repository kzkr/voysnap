package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/kzkr/voysnap/internal/paste"
)

// showResult displays transcribed text in a small window. This is the fallback
// used when Voysnap can't paste (Accessibility not granted); the text is also
// on the clipboard.
func (a *App) showResult(text string) {
	fyne.Do(func() {
		w := a.fyne.NewWindow("Voysnap — Transcript")

		logo := container.NewGridWrap(fyne.NewSize(34, 34), widget.NewIcon(a.iconIdle))
		title := widget.NewRichTextFromMarkdown("## Voysnap")
		header := container.NewHBox(logo, title)

		entry := widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		entry.SetText(text)

		copyBtn := widget.NewButton("Copy", func() { paste.SetClipboard(text) })
		copyBtn.Importance = widget.HighImportance
		closeBtn := widget.NewButton("Close", func() { w.Close() })
		buttons := container.NewGridWithColumns(2, copyBtn, closeBtn)

		content := container.NewBorder(
			container.NewVBox(header, widget.NewSeparator()),
			buttons, nil, nil,
			entry,
		)
		w.SetContent(container.NewPadded(content))
		w.Resize(fyne.NewSize(440, 240))
		w.CenterOnScreen()
		w.Show()
	})
}
