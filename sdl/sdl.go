package sdl

/*
#cgo LDFLAGS: -lSDL2

#include <SDL2/SDL.h>

void Init() {
	SDL_Init(SDL_INIT_VIDEO);
};

void Quit() {
	SDL_Quit();
}

void CreateWindow() {
	SDL_CreateWindow(
		"lol",
		SDL_WINDOWPOS_UNDEFINED,
		SDL_WINDOWPOS_UNDEFINED,
		640,
		480,
		SDL_WINDOW_SHOWN);
}

*/
import "C"

import ()

func Init() {
	C.Init()
}

func Quit() {
	C.Quit()
}

func CreateWindow() {
	C.CreateWindow()
}
