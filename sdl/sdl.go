package sdl

/*
#cgo LDFLAGS: -lSDL2
#include "sdl.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func checkError(cond bool) error {
	var err error
	if cond {
		err = errors.New(C.GoString(C.SDL_GetError()))
	} else {
		err = nil
	}
	return err
}

func Init() error {
	return checkError(int(C.Init()) != 0)
}

func Quit() {
	C.SDL_Quit()
}

type Window *C.SDL_Window

func CreateWindow(title string, w, h int) (Window, error) {
	c_title := C.CString(title)
	defer C.free(unsafe.Pointer(c_title))
	window := C.CreateWindow(c_title, C.int(w), C.int(h))
	err := checkError(C.IsNULL(unsafe.Pointer(window)) != 0)

	return window, err
}

func DestroyWindow(window Window) {
	C.SDL_DestroyWindow(window)
}

type Renderer *C.SDL_Renderer

func CreateRenderer(window Window, index int) (Renderer, error) {
	renderer := C.CreateRenderer(window, C.int(index))
	err := checkError(C.IsNULL(unsafe.Pointer(renderer)) != 0)

	return renderer, err
}

func DestroyRenderer(renderer Renderer) {
	C.SDL_DestroyRenderer(renderer)
}

func RenderClear(renderer Renderer) error {
	return checkError(int(C.SDL_RenderClear(renderer)) != 0)
}

func RenderPresent(renderer Renderer) {
	C.SDL_RenderPresent(renderer)
}

type Texture *C.SDL_Texture

func LoadBMP(renderer Renderer, file string) (Texture, error) {
	c_file := C.CString(file)
	defer C.free(unsafe.Pointer(c_file))
	texture := C.LoadBMP(renderer, c_file)
	err := checkError(C.IsNULL(unsafe.Pointer(texture)) != 0)

	return texture, err
}

func DestroyTexture(texture Texture) {
	C.SDL_DestroyTexture(texture)
}

func RenderCopy(renderer Renderer, texture Texture, x, y, w, h int) error {
	return checkError(int(C.RenderCopy(renderer, texture, C.int(x),
		C.int(y), C.int(w), C.int(h))) != 0)
}

var Running bool = true

type mouse_state struct {
	Down bool
	X    int
	Y    int
}

var Mouse_state mouse_state

func ShowCursor(toggle bool) {
	if toggle {
		C.SDL_ShowCursor(C.int(1))
	} else {
		C.SDL_ShowCursor(C.int(0))
	}
}

func HandleEvents() {
	Mouse_state.Down = false

	for int(C.PollEvent()) != 0 {
		switch C.LastEventType() {
		case C.SDL_QUIT:
			Running = false
		case C.SDL_MOUSEBUTTONDOWN:
			Mouse_state.Down = true
			Mouse_state.X = int(C.MouseX())
			Mouse_state.Y = int(C.MouseY())
		}
	}
}
