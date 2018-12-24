package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/cmplx"
	"os"
)

func main() {
	colorScheme := map[float64]color.RGBA{
		0.0: {0, 0, 0, 255},
		0.1: {139, 113, 201, 255},
		0.2: {122, 164, 68, 255},
		0.3: {102, 40, 80, 255},
		0.5: {77, 173, 152, 255},
		0.7: {197, 120, 62, 255},
		1.0: {0, 0, 0, 255},
	}

	//window := Window{-1.252237535, -0.3451567, 0.0000003, 0.0000003, 1024 * 8, 1024 * 8}
	window := Window{-1.052237535, -0.2503000, 0.0002, 0.0002, 1024 * 8, 1024 * 8}
	parameters := imgParam{window, 1 + 0.0i, 10e6, 10e2, colorScheme}
	img := calcAreaParallel(parameters, 128, 128)

	f, err := os.OpenFile("out.png", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	png.Encode(f, img)
}

type Window struct {
	X1          float64
	Y1          float64
	Width       float64
	Height      float64
	PixelWidth  int
	PixelHeight int
}

type imgParam struct {
	Window        Window
	Z             complex128
	EscapeLimit   float64
	MaxIterations int
	ColorScheme   map[float64]color.RGBA
}

type imgPacket struct {
	Img image.Image
	Px  int
	Py  int
}

func calcAreaParallel(parameters imgParam, xsplits, ysplits int) image.Image {
	pxstep := parameters.Window.PixelWidth / xsplits
	pystep := parameters.Window.PixelHeight / ysplits
	xstep := parameters.Window.Width / float64(xsplits)
	ystep := parameters.Window.Height / float64(ysplits)
	x := parameters.Window.X1
	y := parameters.Window.Y1

	ch := make(chan imgPacket, 255)

	for py := 0; py < parameters.Window.PixelHeight; py += pystep {
		for px := 0; px < parameters.Window.PixelWidth; px += pxstep {
			subwindow := Window{x, y, xstep, ystep, pxstep, pystep}
			if (parameters.Window.PixelWidth - px) < pxstep {
				subwindow.Width = parameters.Window.Width - x
				subwindow.PixelWidth = parameters.Window.PixelWidth - px
			}
			if (parameters.Window.PixelHeight - py) < pystep {
				subwindow.Height = parameters.Window.Height - y
				subwindow.PixelHeight = parameters.Window.PixelHeight - py
			}

			subparameters := imgParam{subwindow, parameters.Z, parameters.EscapeLimit, parameters.MaxIterations, parameters.ColorScheme}
			go calcAreaWrapper(subparameters, px, py, ch)
			x += xstep
		}
		x = parameters.Window.X1
		y += ystep
	}

	img := image.NewRGBA(image.Rect(0, 0, parameters.Window.PixelWidth, parameters.Window.PixelHeight))

	for i := 0; i < ysplits; i++ {
		for j := 0; j < xsplits; j++ {
			subimgPacket := <-ch
			subStart := image.Pt(subimgPacket.Px, subimgPacket.Py)
			subRect := subimgPacket.Img.Bounds()
			subRect.Max = subRect.Max.Add(subStart)
			subRect.Min = subRect.Min.Add(subStart)
			draw.Draw(img, subRect, subimgPacket.Img, image.Point{0, 0}, draw.Src)

			fmt.Printf("%v, %v\t%v, %v\n", i, j, subimgPacket.Px, subimgPacket.Py)
		}
	}

	return img
}

func calcAreaWrapper(params imgParam, px, py int, ch chan imgPacket) {
	ch <- imgPacket{calcArea(params), px, py}
}

func calcArea(parameters imgParam) image.Image {
	window := parameters.Window
	z := parameters.Z
	escapeLimit := parameters.EscapeLimit
	maxIterations := parameters.MaxIterations
	colorScheme := parameters.ColorScheme

	img := image.NewRGBA(image.Rect(0, 0, window.PixelWidth, window.PixelHeight))
	xstep := window.Width / float64(window.PixelWidth)
	ystep := window.Height / float64(window.PixelHeight)

	x := window.X1
	y := window.Y1
	for py := 0; py < window.PixelWidth; py++ {
		for px := 0; px < window.PixelWidth; px++ {
			iterations := calcMandelbrot(complex(x, y), z, escapeLimit, maxIterations)
			img.Set(px, py, floatToColor(colorScheme, math.Sqrt(float64(iterations)/float64(maxIterations))))
			x += xstep
		}
		y += ystep
		x = window.X1
	}

	return img
}

func calcMandelbrot(c, z complex128, escapeLimit float64, maxIterations int) int {
	var iterations int

	for iterations = 0; cmplx.Abs(z) < escapeLimit && iterations < maxIterations; iterations++ {
		z = z*z + c
	}

	return iterations
}

func scaleColor(c color.RGBA, f float64) color.RGBA {
	r := uint8(f * float64(c.R))
	g := uint8(f * float64(c.G))
	b := uint8(f * float64(c.B))
	a := uint8(f * float64(c.A))
	return color.RGBA{r, g, b, a}
}

func addColors(a, b color.RGBA) color.RGBA {
	return color.RGBA{a.R + b.R, a.G + b.G, a.B + b.B, a.A + b.A}
}

func floatToColor(colorStops map[float64]color.RGBA, v float64) color.RGBA {
	var lowerKey, upperKey float64
	upperKey = 1.0
	for k := range colorStops {
		if k > lowerKey && k <= v {
			lowerKey = k
		}
		if k < upperKey && k >= v {
			upperKey = k
		}
	}

	var alpha float64
	if upperKey > lowerKey {
		alpha = (v - lowerKey) / (upperKey - lowerKey)
	} else {
		alpha = 1.0
	}

	col := addColors(scaleColor(colorStops[upperKey], alpha), scaleColor(colorStops[lowerKey], 1.0-alpha))
	return col
}
