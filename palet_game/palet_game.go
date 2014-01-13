package palet_game

import (
	"math"
	"time"
)

type palet struct {
	x  float64
	y  float64
	vx float64
	vy float64
}

func (this *palet) X() float64 {
	return this.x
}

func (this *palet) Y() float64 {
	return this.y
}

func normalize(x, y float64) (float64, float64) {
	norm := math.Sqrt(x*x + y*y)
	return x / norm, y / norm
}

func (this *palet) Launch(dir_x, dir_y int) {
	this.vx, this.vy = normalize(float64(dir_x), float64(dir_y))
	speed := 42.0
	this.vx *= speed
	this.vy *= speed
}

func (this *palet) handleCollision() {
	const (
		palet_radius  = 25.0
		x_left        = 19.0
		x_right       = 420.0
		x_wall_left   = 159.0
		x_wall_right  = 280.0
		y_top         = 19.0
		y_bottom      = 600.0
		y_wall_top    = 290.0
		y_wall_bottom = 329.0
	)

	if this.x-palet_radius <= x_left {
		this.x = x_left + palet_radius
		this.vx *= -1
	} else if this.x+palet_radius >= x_right {
		this.x = x_right - palet_radius
		this.vx *= -1
	}

	if this.y-palet_radius <= y_top {
		this.y = y_top + palet_radius
		this.vy *= -1
	} else if this.y+palet_radius >= y_bottom {
		this.y = y_bottom - palet_radius
		this.vy *= -1
	}
}

type PaletGame struct {
	Palets *[8]palet
}

func NewPaletGame() PaletGame {
	const (
		width  = 440.0
		height = 620.0
	)
	var palets [8]palet

	k := 0
	for i := 0; i < 2; i++ {
		for j := 0; j < 4; j++ {
			palets[k].x = width / 4.0 * float64(1 + 2*i)
			palets[k].y = height / 8.0 * float64(1 + 2*j)
			k++
		}
	}

	return PaletGame{&palets}
}

func (this PaletGame) Step(dt time.Duration) {
	const (
		timestep     = 42.0
		acceleration = -42.0
	)
	accumulator := dt.Seconds()

	for accumulator >= timestep {
		for i := range this.Palets {
			this.Palets[i].x += this.Palets[i].vx * timestep
			this.Palets[i].y += this.Palets[i].vy * timestep

			dir_x, dir_y := normalize(this.Palets[i].vx,
				this.Palets[i].vy)

			this.Palets[i].vx += dir_x * acceleration * timestep
			if this.Palets[i].vx < 0 {
				this.Palets[i].vx = 0
			}
			this.Palets[i].vy += dir_y * acceleration * timestep
			if this.Palets[i].vy < 0 {
				this.Palets[i].vy = 0
			}
		}
	}
}
