package sdl

/*
#cgo LDFLAGS: -lSDL2
#include "sdl.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

func Init() {
	C.Init()
}

func Quit() {
	C.Quit()
}

type Window *C.SDL_Window

func CreateWindow(title string, w, h int) Window {
	c_title := C.CString(title)
	defer C.free(unsafe.Pointer(c_title))
	return C.CreateWindow(c_title, C.int(w), C.int(h))
}

func DestroyWindow(window Window) {
	C.SDL_DestroyWindow(window)
}

type Renderer *C.SDL_Renderer

func CreateRenderer(window Window, index int) Renderer {
	return C.CreateRenderer(window, C.int(index))
}

func DestroyRenderer(renderer Renderer) {
	C.SDL_DestroyRenderer(renderer)
}

func RenderPresent(renderer Renderer) {
	C.SDL_RenderPresent(renderer)
}

type Texture *C.SDL_Texture

func LoadBMP(renderer Renderer, file string) Texture {
	c_file := C.CString(file)
	defer C.free(unsafe.Pointer(c_file))
	return C.LoadBMP(renderer, c_file)
}

func DestroyTexture(texture Texture) {
	C.SDL_DestroyTexture(texture)
}

func PollEvent() int {
	return int(C.PollEvent())
}

func IsLastEventQUIT() bool {
	return C.IsLastEventQUIT() != 0
}
