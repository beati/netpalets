#ifndef SDL_H
#define SDL_H

#include <SDL2/SDL.h>

void Init();

void Quit();

SDL_Window *CreateWindow(const char *title, int w, int h);

SDL_Renderer *CreateRenderer(SDL_Window *window, int index);

int PollEvent();

int IsLastEventQUIT();

SDL_Texture *LoadBMP(SDL_Renderer *renderer, const char *file);

#endif
