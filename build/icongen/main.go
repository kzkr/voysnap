// Command icongen renders Voysnap's icons from the exact "S" monogram defined
// in build/logo-source.svg (four equal squares with 180° rotational symmetry):
//
//	app       → 1024×1024 white rounded square + black S (for icon.icns)
//	bar-idle  → 256×256 transparent black S, used as a template (menu-bar, idle)
//	bar-rec   → 256×256 transparent red S                     (menu-bar, recording)
//
// Usage: go run ./build/icongen <mode> <out.png>
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// The four squares of the S monogram, in the source SVG's 2048-unit space.
var sBlocks = [][4]float64{
	{1024, 470, 1393, 839.333},   // top-right
	{655, 654.667, 1024, 1024},   // upper-left
	{1024, 1024, 1393, 1393.333}, // lower-right
	{655, 1208.667, 1024, 1578},  // bottom-left
}

// Bounding box of the monogram (taller than wide), centred at (1024, 1024).
const (
	logoCX = 1024.0
	logoCY = 1024.0
	logoH  = 1578 - 470 // 1108
)

func main() {
	mode, out := os.Args[1], os.Args[2]
	switch mode {
	case "app":
		renderApp(out)
	case "bar-idle":
		renderBar(out, color.NRGBA{0x00, 0x00, 0x00, 0xff}) // colour ignored when used as a template
	case "bar-rec":
		renderBar(out, color.NRGBA{0xe5, 0x3a, 0x3a, 0xff})
	default:
		panic("unknown mode: " + mode)
	}
}

// renderApp draws the app icon: a black S on a white rounded square.
func renderApp(out string) {
	const size = 1024
	const logoFill = 0.52 // fraction of the canvas the logo's height spans
	c := float64(size) / 2
	scale := logoFill * size / logoH

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5

			bg := sdRoundBox(px-c, py-c, 430, 430, 200)
			bgA := clamp(0.5-bg, 0, 1)
			if bgA <= 0 {
				continue
			}

			lx := logoCX + (px-c)/scale
			ly := logoCY + (py-c)/scale
			logoA := clamp(0.5-sLogoSDF(lx, ly)*scale, 0, 1)

			v := lerp8(0xFF, 0x0A, logoA) // black S over white
			img.SetNRGBA(x, y, color.NRGBA{v, v, v, uint8(bgA * 255)})
		}
	}
	write(out, img)
}

// renderBar draws the S glyph in col on a transparent canvas, scaled to nearly
// fill it (so it matches the size of other menu-bar icons once macOS scales it).
func renderBar(out string, col color.NRGBA) {
	const size = 256
	const logoFill = 0.86
	c := float64(size) / 2
	scale := logoFill * size / logoH

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			lx := logoCX + (float64(x)+0.5-c)/scale
			ly := logoCY + (float64(y)+0.5-c)/scale
			a := clamp(0.5-sLogoSDF(lx, ly)*scale, 0, 1)
			if a <= 0 {
				continue
			}
			img.SetNRGBA(x, y, color.NRGBA{col.R, col.G, col.B, uint8(a * float64(col.A))})
		}
	}
	write(out, img)
}

// sLogoSDF: signed distance to the union of the four squares.
func sLogoSDF(px, py float64) float64 {
	d := math.MaxFloat64
	for _, b := range sBlocks {
		cx, cy := (b[0]+b[2])/2, (b[1]+b[3])/2
		hx, hy := (b[2]-b[0])/2, (b[3]-b[1])/2
		d = math.Min(d, sdBox(px-cx, py-cy, hx, hy))
	}
	return d
}

func write(out string, img image.Image) {
	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
func lerp8(a, b, t float64) uint8     { return uint8(clamp(a+(b-a)*t, 0, 255)) }

// sdBox: signed distance to an axis-aligned box centred at origin.
func sdBox(px, py, hx, hy float64) float64 {
	dx, dy := math.Abs(px)-hx, math.Abs(py)-hy
	return math.Hypot(math.Max(dx, 0), math.Max(dy, 0)) + math.Min(math.Max(dx, dy), 0)
}

// sdRoundBox: signed distance to a rounded box centred at origin.
func sdRoundBox(px, py, hx, hy, r float64) float64 {
	qx := math.Abs(px) - hx + r
	qy := math.Abs(py) - hy + r
	return math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - r
}
