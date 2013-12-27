#ifndef SDL_H
#define SDL_H

#include <SDL2/SDL.h>

int IsNULL(void *p);

int Init();

SDL_Window *CreateWindow(const char *title, int w, int h);

SDL_Renderer *CreateRenderer(SDL_Window *window, int index);

int PollEvent();

int IsLastEventQUIT();

SDL_Texture *LoadBMP(SDL_Renderer *renderer, const char *file);

int RenderCopy(SDL_Renderer *renderer, SDL_Texture *texture);

int RenderCopyXY(SDL_Renderer *renderer, SDL_Texture *texture, int x, int y,
		int w, int h);

#endif
