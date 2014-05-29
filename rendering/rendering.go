package rendering

import (
	"bytes"
	"encoding/binary"
	"github.com/beati/netpalets/gamestate"
	"github.com/beati/netpalets/sdl"
	"log"
)

var window sdl.Window
var renderer sdl.Renderer
var board_gfx sdl.Texture
var palet_gfx sdl.Texture

func InitRendering() {
	var err error

	window, err = sdl.CreateWindow("foo", 440, 620)
	if err != nil {
		log.Fatal(err)
	}

	renderer, err = sdl.CreateRenderer(window, -1)
	if err != nil {
		log.Fatal(err)
	}

	board_gfx, err = sdl.LoadBMP(renderer, "board.bmp")
	if err != nil {
		log.Fatal(err)
	}

	palet_gfx, err = sdl.LoadBMP(renderer, "palet.bmp")
	if err != nil {
		log.Fatal(err)
	}
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

func RenderFromNet(gameState []byte) {
	var err error

	err = sdl.RenderClear(renderer)
	if err != nil {
		log.Fatal(err)
	}

	err = sdl.RenderCopy(renderer, board_gfx, 0, 0, 440, 620)
	if err != nil {
		log.Fatal(err)
	}

	r := bytes.NewReader(gameState)
	for i := 0; i < 8; i++ {
		var xf float64
		var yf float64
		binary.Read(r, binary.LittleEndian, &xf)
		binary.Read(r, binary.LittleEndian, &yf)
		x := shiftPos(xf)
		y := shiftPos(yf)
		err = sdl.RenderCopy(renderer, palet_gfx, x, y, 50, 50)
		if err != nil {
			log.Fatal(err)
		}
	}

	sdl.RenderPresent(renderer)
}

func Render(gameState *gamestate.GameState) {
	var err error

	err = sdl.RenderClear(renderer)
	if err != nil {
		log.Fatal(err)
	}

	err = sdl.RenderCopy(renderer, board_gfx, 0, 0, 440, 620)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 8; i++ {
		x := shiftPos(gameState.X(i))
		y := shiftPos(gameState.Y(i))
		err = sdl.RenderCopy(renderer, palet_gfx, x, y, 50, 50)
		if err != nil {
			log.Fatal(err)
		}
	}

	sdl.RenderPresent(renderer)
}
