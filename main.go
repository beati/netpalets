package main

import (
	"github.com/beati/NetPalets/sdl"
)

func main() {
	sdl.Init()
	defer sdl.Quit()

	window := sdl.CreateWindow("foo", 300, 200)
	defer sdl.DestroyWindow(window)

	renderer := sdl.CreateRenderer(window, -1)
	defer sdl.DestroyRenderer(renderer)

	for running := true; running; {
		for sdl.PollEvent() != 0 {
			if sdl.IsLastEventQUIT() {
				running = false
			}
		}
		sdl.RenderPresent(renderer)
	}
}
