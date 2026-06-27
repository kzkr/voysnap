// Command icongen renders VoySnap's icons from the exact logo defined in
// build/logo-source.svg (a left square joined to a curved "wave"):
//
//	app       → 1024×1024 white rounded square + black logo (for icon.icns)
//	bar-idle  → 256×256 transparent black logo, used as a template (menu-bar, idle)
//	bar-rec   → 256×256 transparent red logo                     (menu-bar, recording)
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

// Logo geometry, in the source SVG's 2048-unit space. The logo is the union of:
//   - a left square (logoRect), and
//   - a wave to its right (wavePoly), bounded by two cubic béziers and two
//     vertical edges, flattened to a polygon at init.
//
// Square: M533 562 H1006.04 V1033.99 H533 V562 Z
var logoRect = [4]float64{533, 562, 1006.04, 1033.99} // x0,y0,x1,y1

// Wave: M1006.04 1033.99
//
//	C1125.79 1033.99 1275.49 562 1515 562   (top edge)
//	V1034                                    (right edge)
//	C1395.24 1034 1245.55 1506 1006.04 1506  (bottom edge)
//	V1033.99 Z                               (left edge)
var wavePoly [][2]float64

func init() {
	pts := [][2]float64{{1006.04, 1033.99}}
	appendCubic(&pts, 1006.04, 1033.99, 1125.79, 1033.99, 1275.49, 562, 1515, 562)
	pts = append(pts, [2]float64{1515, 1034})
	appendCubic(&pts, 1515, 1034, 1395.24, 1034, 1245.55, 1506, 1006.04, 1506)
	pts = append(pts, [2]float64{1006.04, 1033.99})
	wavePoly = pts
}

// Bounding box of the whole logo, and its centre, in SVG units.
const (
	logoMinX = 533.0
	logoMaxX = 1515.0
	logoMinY = 562.0
	logoMaxY = 1506.0
	logoCX   = (logoMinX + logoMaxX) / 2 // 1024
	logoCY   = (logoMinY + logoMaxY) / 2 // 1034
	logoSpan = logoMaxX - logoMinX       // 982 (the larger dimension; wider than tall)
)

// ss is the per-axis supersampling factor used to antialias the logo edges.
const ss = 4

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

// renderApp draws the app icon: a black logo on a white rounded square.
func renderApp(out string) {
	const size = 1024
	const logoFill = 0.52 // fraction of the canvas the logo's larger side spans
	c := float64(size) / 2
	scale := logoFill * size / logoSpan

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			bg := sdRoundBox(float64(x)+0.5-c, float64(y)+0.5-c, 430, 430, 200)
			bgA := clamp(0.5-bg, 0, 1)
			if bgA <= 0 {
				continue
			}
			logoA := logoCoverage(float64(x), float64(y), c, scale)
			v := lerp8(0xFF, 0x0A, logoA) // black logo over white
			img.SetNRGBA(x, y, color.NRGBA{v, v, v, uint8(bgA * 255)})
		}
	}
	write(out, img)
}

// renderBar draws the logo in col on a transparent canvas, scaled to nearly
// fill it (so it matches the size of other menu-bar icons once macOS scales it).
func renderBar(out string, col color.NRGBA) {
	const size = 256
	const logoFill = 0.86
	c := float64(size) / 2
	scale := logoFill * size / logoSpan

	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			a := logoCoverage(float64(x), float64(y), c, scale)
			if a <= 0 {
				continue
			}
			img.SetNRGBA(x, y, color.NRGBA{col.R, col.G, col.B, uint8(a * float64(col.A))})
		}
	}
	write(out, img)
}

// logoCoverage returns the fraction (0..1) of pixel (x,y) covered by the logo,
// estimated by ss×ss supersampling. c is the canvas centre and scale maps SVG
// units to canvas pixels (logo centred at the canvas centre).
func logoCoverage(x, y, c, scale float64) float64 {
	hits := 0
	for sy := 0; sy < ss; sy++ {
		for sx := 0; sx < ss; sx++ {
			px := x + (float64(sx)+0.5)/ss
			py := y + (float64(sy)+0.5)/ss
			lx := logoCX + (px-c)/scale
			ly := logoCY + (py-c)/scale
			if insideLogo(lx, ly) {
				hits++
			}
		}
	}
	return float64(hits) / float64(ss*ss)
}

// insideLogo reports whether SVG-space point (x,y) is inside the logo.
func insideLogo(x, y float64) bool {
	if x >= logoRect[0] && x <= logoRect[2] && y >= logoRect[1] && y <= logoRect[3] {
		return true
	}
	return pointInPoly(x, y, wavePoly)
}

// pointInPoly is a standard even-odd ray-cast point-in-polygon test.
func pointInPoly(x, y float64, poly [][2]float64) bool {
	in := false
	n := len(poly)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := poly[i][0], poly[i][1]
		xj, yj := poly[j][0], poly[j][1]
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}

// appendCubic flattens a cubic bézier into line segments, appending the points
// after the start (the start is assumed already present in dst).
func appendCubic(dst *[][2]float64, x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	const steps = 64
	for i := 1; i <= steps; i++ {
		t := float64(i) / steps
		u := 1 - t
		a, b, c, d := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
		px := a*x0 + b*x1 + c*x2 + d*x3
		py := a*y0 + b*y1 + c*y2 + d*y3
		*dst = append(*dst, [2]float64{px, py})
	}
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

// sdRoundBox: signed distance to a rounded box centred at origin.
func sdRoundBox(px, py, hx, hy, r float64) float64 {
	qx := math.Abs(px) - hx + r
	qy := math.Abs(py) - hy + r
	return math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - r
}
