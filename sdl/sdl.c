#include "sdl.h"

int IsNULL(void *p) {
	return p == NULL;
}

int Init() {
	return SDL_Init(SDL_INIT_VIDEO);
};

SDL_Window *CreateWindow(const char *title, int w, int h) {
	return SDL_CreateWindow(
			title,
			SDL_WINDOWPOS_UNDEFINED,
			SDL_WINDOWPOS_UNDEFINED,
			w,
			h,
			0);
}

SDL_Renderer *CreateRenderer(SDL_Window *window, int index) {
	return SDL_CreateRenderer(window, index, SDL_RENDERER_PRESENTVSYNC);
}

SDL_Texture *LoadBMP(SDL_Renderer *renderer, const char *file) {
	SDL_Surface *image = SDL_LoadBMP(file);
	if (image == NULL) return NULL;
	SDL_Texture *texture = SDL_CreateTextureFromSurface(renderer, image);
	SDL_FreeSurface(image);
	return texture;
}

int RenderCopy(SDL_Renderer *renderer, SDL_Texture *texture) {
	return SDL_RenderCopy(renderer, texture, NULL, NULL);
}

int RenderCopyXY(SDL_Renderer *renderer, SDL_Texture *texture, int x, int y,
		int w, int h) {
	SDL_Rect dst;
	dst.x = x;
	dst.y = y;
	dst.w = w;
	dst.h = h;
	return SDL_RenderCopy(renderer, texture, NULL, &dst);
}

static SDL_Event last_event;

int PollEvent() {
	return SDL_PollEvent(&last_event);
}

int IsLastEventQUIT() {
	return last_event.type == SDL_QUIT;
}
