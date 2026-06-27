package app

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// brandColor is the black from the SilentRec logo, used as the accent instead of
// Fyne's default blue.
var brandColor = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}

// silentTheme forces a light (white) appearance with a black accent, regardless
// of the system's dark mode, so the windows match the logo's look.
type silentTheme struct{}

var _ fyne.Theme = silentTheme{}

func (silentTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return brandColor
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0xff} // pure black text
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // pure white window
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0x2a}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0x40}
	}
	// Everything else uses the built-in light palette. (HighImportance button
	// text uses ForegroundOnPrimary, left at default, so Save stays white on black.)
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (silentTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (silentTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (silentTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }
