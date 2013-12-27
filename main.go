package main

import (
	"github.com/beati/NetPalets/sdl"
	"log"
)

func main() {
	var err error

	err = sdl.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("foo", 320, 240)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyWindow(window)

	renderer, err := sdl.CreateRenderer(window, -1)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyRenderer(renderer)

	image, err := sdl.LoadBMP(renderer, "sdl_logo.bmp")
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyTexture(image)

	err = sdl.RenderCopy(renderer, image)
	if err != nil {
		log.Fatal(err)
	}

	err = sdl.RenderCopyXY(renderer, image, 40, 40, 50, 57)
	if err != nil {
		log.Fatal(err)
	}

	for running := true; running; {
		for sdl.PollEvent() {
			if sdl.IsLastEventQUIT() {
				running = false
			}
		}
		sdl.RenderPresent(renderer)
	}
}
