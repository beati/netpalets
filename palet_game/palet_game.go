package palet_game

import (
	"time"
)

type palet struct {
	int x
	int y
	int vx
	int vy
}

func (this palet) X() {
}

func (this palet) Y() {
}

func (this palet) Launch( /* direction */) {
}

type PaletGame struct {
	palets *[8]palet
}

func (this PaletGame) Step(dt time.Duration) {
}
