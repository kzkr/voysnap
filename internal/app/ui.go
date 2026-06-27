package app

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/kzkr/silentrec/internal/paste"
)

// header builds a logo + title row reused across windows. The logo uses the
// idle (themed) icon, so it adapts to light/dark like the menu-bar icon.
func (a *App) header(subtitle string) fyne.CanvasObject {
	logo := container.NewGridWrap(fyne.NewSize(34, 34), widget.NewIcon(a.iconIdle))

	title := widget.NewRichTextFromMarkdown("## SilentRec")
	title.Wrapping = fyne.TextWrapOff

	titleCol := container.NewVBox(title)
	if subtitle != "" {
		titleCol.Add(widget.NewLabel(subtitle))
	}
	return container.NewHBox(logo, titleCol)
}

// showResult displays transcribed text in a small window (the no-auto-paste
// fallback). The text is already on the clipboard.
func (a *App) showResult(text string) {
	fyne.Do(func() {
		w := a.fyne.NewWindow("SilentRec — Transcript")

		entry := widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		entry.SetText(text)

		copyBtn := widget.NewButtonWithIcon("Copy", nil, func() { paste.SetClipboard(text) })
		copyBtn.Importance = widget.HighImportance
		closeBtn := widget.NewButton("Close", func() { w.Close() })
		buttons := container.NewGridWithColumns(2, copyBtn, closeBtn)

		content := container.NewBorder(
			container.NewVBox(a.header(""), widget.NewSeparator()),
			buttons, nil, nil,
			entry,
		)
		w.SetContent(container.NewPadded(content))
		w.Resize(fyne.NewSize(440, 260))
		w.CenterOnScreen()
		w.Show()
	})
}

// showSettings opens (or focuses) the settings window.
func (a *App) showSettings() {
	a.mu.Lock()
	if a.settingsOpen {
		a.mu.Unlock()
		return
	}
	a.settingsOpen = true
	a.mu.Unlock()

	fyne.Do(func() {
		w := a.fyne.NewWindow("SilentRec — Settings")
		closeSettings := func() {
			a.mu.Lock()
			a.settingsOpen = false
			a.mu.Unlock()
			w.Close()
		}
		w.SetCloseIntercept(closeSettings)

		modelEntry := widget.NewEntry()
		modelEntry.SetText(a.cfg.ModelPath)

		browse := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
			d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
				if err != nil || r == nil {
					return
				}
				defer r.Close()
				modelEntry.SetText(r.URI().Path())
			}, w)
			d.SetFilter(storage.NewExtensionFileFilter([]string{".bin"}))
			d.Resize(fyne.NewSize(700, 500))
			d.Show()
		})
		modelRow := container.NewBorder(nil, nil, nil, browse, modelEntry)

		langSelect := widget.NewSelect(langNames, nil)
		langSelect.SetSelectedIndex(langIndex(a.cfg.Language))

		vocab := widget.NewMultiLineEntry()
		vocab.SetText(strings.Join(a.cfg.Vocabulary, ", "))
		vocab.SetMinRowsVisible(2)

		snip := widget.NewMultiLineEntry()
		snip.SetText(formatSnippets(a.cfg.Snippets))
		snip.SetMinRowsVisible(3)

		modelItem := widget.NewFormItem("Model", modelRow)
		langItem := widget.NewFormItem("Language", langSelect)
		vocabItem := widget.NewFormItem("Custom words", vocab)
		vocabItem.HintText = "Names/jargon to recognize better, comma-separated"
		snipItem := widget.NewFormItem("Snippets", snip)
		snipItem.HintText = "One per line:  phrase = replacement"
		form := widget.NewForm(modelItem, langItem, vocabItem, snipItem)

		save := widget.NewButton("Save", func() {
			a.cfg.ModelPath = strings.TrimSpace(modelEntry.Text)
			a.cfg.Language = langCodes[langSelect.SelectedIndex()]
			a.cfg.Vocabulary = parseList(vocab.Text)
			a.cfg.Snippets = parseSnippets(snip.Text)
			if err := a.cfg.Save(); err != nil {
				dialog.ShowError(err, w)
				return
			}
			closeSettings()
		})
		save.Importance = widget.HighImportance

		hotkeyNote := widget.NewLabel("Tap the right ⌘ key to start or stop dictation.")

		body := container.NewVBox(
			a.header("Local, offline voice dictation"),
			widget.NewSeparator(),
			hotkeyNote,
			form,
			widget.NewSeparator(),
			save,
		)
		w.SetContent(container.NewPadded(body))
		w.Resize(fyne.NewSize(500, 460))
		w.CenterOnScreen()
		w.Show()
	})
}

// Language options shown in settings; langCodes[i] is the whisper code for
// langNames[i] ("auto" = auto-detect).
var (
	langNames = []string{"Auto-detect", "English", "French", "Spanish", "German", "Italian", "Portuguese", "Dutch"}
	langCodes = []string{"auto", "en", "fr", "es", "de", "it", "pt", "nl"}
)

func langIndex(code string) int {
	for i, c := range langCodes {
		if c == code {
			return i
		}
	}
	return 0 // Auto-detect
}

// parseList splits a comma/newline-separated string into trimmed, non-empty items.
func parseList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// formatSnippets renders a snippet map as "phrase = replacement" lines (sorted
// for stable display).
func formatSnippets(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(" = ")
		b.WriteString(m[k])
		b.WriteByte('\n')
	}
	return b.String()
}

// parseSnippets parses "phrase = replacement" lines into a map.
func parseSnippets(s string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		if key != "" && val != "" {
			out[key] = val
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
