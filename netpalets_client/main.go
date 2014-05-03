package main

import (
	"github.com/beati/netpalets/fatal"
	"github.com/beati/netpalets/netpalets_client/rendering"
	"github.com/beati/netpalets/sdl"
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

	t := time.Now()

	for sdl.Running {
		dt := time.Since(t)
		t = time.Now()
		if sdl.Mouse_state.Down {
			game_state.Palets[0].Launch(sdl.Mouse_state.X,
				sdl.Mouse_state.Y)
		}
		game_state.Step(dt)
		sdl.HandleEvents()
		rendering.Render(game_state)
	}
}
