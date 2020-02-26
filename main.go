package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/fogleman/gg"
	"image"
	"math/cmplx"
)

type Bounds struct {
	X_min float64
	Y_min float64
	X_max float64
	Y_max float64
}

type Color struct {
	R float64
	G float64
	B float64
}

type empty struct{}
type semaphore chan empty

func run(S int, pixelBounds, renderBounds Bounds) func() {
	return func() {
		cfg := pixelgl.WindowConfig{
			Title:  "Mandelbrot",
			Bounds: pixel.R(0, 0, float64(S), float64(S)),
			VSync:  true,
		}
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			panic(err)
		}

		img := renderFrame(pixelBounds, renderBounds)
		pic := pixel.PictureDataFromImage(img)
		sprite := pixel.NewSprite(pic, pic.Bounds())

		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

		for !win.Closed() {
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				click := win.MousePosition()
				newX, newY := mapBounds(click.X, float64(S)-click.Y, pixelBounds, renderBounds)

				renderBounds = pointZoomToBounds(newX, newY, 0.5*(renderBounds.X_max-renderBounds.X_min))

				img = renderFrame(pixelBounds, renderBounds)
				pic = pixel.PictureDataFromImage(img)
				sprite = pixel.NewSprite(pic, pic.Bounds())

				sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
				fmt.Println(click, newX, newY)
			}

			win.Update()
		}
	}
}

func main() {
	const S = 512
	pixelBounds := Bounds{0, 0, float64(S), float64(S)}
	//renderBounds := pointZoomToBounds(-0.5, 0.620, 0.01)
	renderBounds := pointZoomToBounds(0, 0, 2)

	pixelgl.Run(run(S, pixelBounds, renderBounds))
}

func renderFrame(pixelBounds, renderBounds Bounds) image.Image {
	S := int(pixelBounds.X_max - pixelBounds.X_min)
	dc := gg.NewContext(S, S)

	colors := []Color{
		{0, 0, 0},
		{0.4039, 0.1608, 0.5137},
		{0.0372, 0.8059, 0.5275},
		{0.1023, 0.3120, 0.7021},
		{0.2372, 0.9059, 0.7275},
	}

	stops := []float64{0.0, 0.05, 0.2, 0.4, 1.0}

	for x_ := 0; x_ < S; x_++ {
		for y_ := 0; y_ < S; y_++ {
			x, y := mapBounds(float64(x_), float64(y_), pixelBounds, renderBounds)

			v := inSet(complex(x, y), 500, 1000000.0)
			c := mapColor(v, colors, stops)

			dc.SetRGB(c.R, c.G, c.B)
			dc.SetPixel(x_, y_)
		}
	}

	return dc.Image()
}

func pointZoomToBounds(x, y, zoom float64) Bounds {
	return Bounds{x - 0.5*zoom, y - 0.5*zoom, x + 0.5*zoom, y + 0.5*zoom}
}

func inSet(a complex128, maxIterations int, boundary float64) float64 {
	iter := 0
	a_i := a
	for ; iter < maxIterations && cmplx.Abs(a_i) < boundary; iter++ {
		a_i = a_i*a_i + a
	}

	if iter >= maxIterations {
		return 0.0
	}

	return float64(iter) / float64(maxIterations)
}

func mapBounds(x_a, y_a float64, a_bounds, b_bounds Bounds) (float64, float64) {
	x_delta := (x_a - a_bounds.X_min) / (a_bounds.X_max - a_bounds.X_min)
	y_delta := (y_a - a_bounds.Y_min) / (a_bounds.Y_max - a_bounds.Y_min)
	x_b := x_delta*(b_bounds.X_max-b_bounds.X_min) + b_bounds.X_min
	y_b := y_delta*(b_bounds.Y_max-b_bounds.Y_min) + b_bounds.Y_min

	return x_b, y_b
}

func (c Color) scale(v float64) Color {
	return Color{c.R * v, c.G * v, c.B * v}
}

func (c Color) add(b Color) Color {
	return Color{c.R + b.R, c.G + b.G, c.B + b.B}
}

func mapColor(v float64, colors []Color, stops []float64) Color {
	for i := range stops {
		if v < stops[i] {
			v_delta := (v - stops[i-1]) / (stops[i] - stops[i-1])
			return colors[i].scale(v_delta).add(colors[i-1].scale(1.0 - v_delta))
		}
	}

	return Color{0, 0, 0}
}
