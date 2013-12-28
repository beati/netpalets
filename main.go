package main

import (
	"github.com/beati/NetPalets/fatal"
	"github.com/beati/NetPalets/sdl"
)

func main() {
	var err error

	err = sdl.Init()
	fatal.Check(err)
	defer sdl.Quit()

	window, err := sdl.CreateWindow("foo", 320, 240)
	fatal.Check(err)
	defer sdl.DestroyWindow(window)

	renderer, err := sdl.CreateRenderer(window, -1)
	fatal.Check(err)
	defer sdl.DestroyRenderer(renderer)

	image, err := sdl.LoadBMP(renderer, "sdl_logo.bmp")
	fatal.Check(err)
	defer sdl.DestroyTexture(image)

	sdl.ShowCursor(false)

	for sdl.Running {
		sdl.HandleEvents()

		err = sdl.RenderClear(renderer)
		fatal.Check(err)

		err = sdl.RenderCopy(renderer, image, 0, 0, 320, 240)
		fatal.Check(err)

		err = sdl.RenderCopy(renderer, image, sdl.Mouse_state.X,
			sdl.Mouse_state.Y, 50, 57)
		fatal.Check(err)

		sdl.RenderPresent(renderer)
	}
}
