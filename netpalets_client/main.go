package main

import (
	"github.com/beati/netpalets/fatal"
	"github.com/beati/netpalets/netpalets_client/rendering"
	"github.com/beati/netpalets/netpalets_client/sdl"
	"github.com/beati/netpalets/palet_game"
)

func main() {
	var err error

	err = sdl.Init()
	fatal.Check(err)
	defer sdl.Quit()

	game_state := palet_game.NewPaletGame()

	rendering.InitRendering()
	defer rendering.CloseRendering()

	//sdl.ShowCursor(false)

	for sdl.Running {
		sdl.HandleEvents()
		rendering.Render(game_state)
	}
}
