package main

import (
	"github.com/beati/netpalets/gamestate"
	"github.com/beati/netpalets/rendering"
	"github.com/beati/netpalets/sdl"
	"log"
	"runtime"
	"time"
)

func main() {
	var err error

	runtime.GOMAXPROCS(4)

	err = sdl.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	gameState := gamestate.NewGameState()

	rendering.InitRendering()
	defer rendering.CloseRendering()

	//sdl.ShowCursor(false)

	t := time.Now()

	for sdl.Running {
		rendering.Render(gameState)

		sdl.HandleEvents()
		if sdl.Mouse.Down {
			gameState.Launch(0, sdl.Mouse.X, sdl.Mouse.Y)
		}

		dt := time.Since(t)
		t = time.Now()
		gameState.Step(dt)
	}
}
