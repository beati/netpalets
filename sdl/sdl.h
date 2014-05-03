#ifndef SDL_H
#define SDL_H

#include <SDL2/SDL.h>

int IsNULL(void *p);

int Init();

SDL_Window *CreateWindow(const char *title, int w, int h);

SDL_Renderer *CreateRenderer(SDL_Window *window, int index);

SDL_Texture *LoadBMP(SDL_Renderer *renderer, const char *file);

int RenderCopy(SDL_Renderer *renderer, SDL_Texture *texture, int x, int y,
		int w, int h);

int PollEvent();

unsigned int event_SDL_QUIT();

unsigned int event_SDL_MOUSEMOTION();

unsigned int LastEventType();

int MouseX();

int MouseY();

int MouseXrel();

int MouseYrel();

#endif
