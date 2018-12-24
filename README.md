# Mandelbrot Fractal
The purpose of this project was to have a little fun with Go's claim to fame — it's goroutines.
It splits the image into n by m subimages, each of which is calculated by a different goroutine.
Once each goroutine is done, it's results are added into the final image. I've found that an
ideal amount of subimages seems to be about 128 by 128, or 16384 pieces, at least for my laptop.
This allows each core to continue working practically until the end.
