package main

import (
	"github.com/beati/netpalets/fatal"
	"github.com/beati/netpalets/netpalets_client/rendering"
	"github.com/beati/netpalets/netpalets_client/sdl"
	"github.com/beati/netpalets/palet_game"
	"runtime"
	"time"
)

func main() {
	var err error

	runtime.GOMAXPROCS(4)

	err = sdl.Init()
	fatal.Check(err)
	defer sdl.Quit()

	game_state := palet_game.NewPaletGame()

	rendering.InitRendering()
	defer rendering.CloseRendering()

	//sdl.ShowCursor(false)

	game_state.Palets[0].Launch(1, 1)

	t := time.Now()

	for sdl.Running {
		dt := time.Since(t)
		t = time.Now()
		game_state.Step(dt)
		sdl.HandleEvents()
		rendering.Render(game_state)
	}
}
