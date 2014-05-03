package rendering

import (
	"github.com/beati/netpalets/fatal"
	"github.com/beati/netpalets/sdl"
	"github.com/beati/netpalets/palet_game"
)

var window sdl.Window
var renderer sdl.Renderer
var board_gfx sdl.Texture
var palet_gfx sdl.Texture

func InitRendering() {
	var err error

	window, err = sdl.CreateWindow("foo", 440, 620)
	fatal.Check(err)

	renderer, err = sdl.CreateRenderer(window, -1)
	fatal.Check(err)

	board_gfx, err = sdl.LoadBMP(renderer, "board.bmp")
	fatal.Check(err)

	palet_gfx, err = sdl.LoadBMP(renderer, "palet.bmp")
	fatal.Check(err)
}

func CloseRendering() {
	sdl.DestroyWindow(window)
	sdl.DestroyRenderer(renderer)
	sdl.DestroyTexture(board_gfx)
	sdl.DestroyTexture(palet_gfx)
}

func shiftPos(x float64) int {
	return int(x+0.5) - 25
}

func Render(game_state palet_game.PaletGame) {
	var err error

	err = sdl.RenderClear(renderer)
	fatal.Check(err)

	err = sdl.RenderCopy(renderer, board_gfx, 0, 0, 440, 620)
	fatal.Check(err)

	for _, palet := range game_state.Palets {
		err = sdl.RenderCopy(renderer, palet_gfx,
			shiftPos(palet.X()), shiftPos(palet.Y()), 50, 50)
		fatal.Check(err)
	}

	sdl.RenderPresent(renderer)
}
