standard: mandelbrot.go
	time go build projects/mandelbrot && time ./mandelbrot && xdg-open out.png

debug: mandelbrot.go
	go build -gcflags=-N projects/mandelbrot && gdb mandelbrot
