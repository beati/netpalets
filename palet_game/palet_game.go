package palet_game

import (
	"github.com/beati/netpalets/fatal"
	"math"
	"time"
)

type palet struct {
	x     float64
	y     float64
	v     float64
	dir_x float64
	dir_y float64
}

func (this *palet) X() float64 {
	return this.x
}

func (this *palet) Y() float64 {
	return this.y
}

func normalize(x, y float64) (float64, float64, float64) {
	norm := math.Sqrt(x*x + y*y)
	return x / norm, y / norm, norm
}

func (this *palet) Launch(dir_x, dir_y int) {
	this.dir_x, this.dir_y, _ = normalize(float64(dir_x)-this.x,
		float64(dir_y)-this.y)
	this.v = 300.0
}

const palet_radius = 25.0

func (this *palet) handleBoardCollision() {
	const (
		x_left        = 19.0
		x_right       = 420.0
		x_wall_left   = 159.0
		x_wall_right  = 280.0
		y_top         = 19.0
		y_bottom      = 600.0
		y_mid         = 310.0
		y_wall_top    = 290.0
		y_wall_bottom = 329.0
	)

	if this.x-palet_radius <= x_left {
		this.x = x_left + palet_radius
		this.dir_x *= -1
	} else if this.x+palet_radius >= x_right {
		this.x = x_right - palet_radius
		this.dir_x *= -1
	}

	if this.y-palet_radius <= y_top {
		this.y = y_top + palet_radius
		this.dir_y *= -1
	} else if this.y+palet_radius >= y_bottom {
		this.y = y_bottom - palet_radius
		this.dir_y *= -1
	}

	if this.x < x_wall_left || this.x > x_wall_right {
		if this.y < y_mid {
			if this.y+palet_radius >= y_wall_top {
				this.y = y_wall_top - palet_radius
				this.dir_y *= -1
			}
		} else {
			if this.y-palet_radius <= y_wall_bottom {
				this.y = y_wall_bottom + palet_radius
				this.dir_y *= -1
			}
		}
	}

	if this.y > y_wall_top && this.y < y_wall_bottom {
		if this.x-palet_radius <= x_wall_left {
			this.x = x_wall_left + palet_radius
			this.dir_x *= -1
		} else if this.x+palet_radius >= x_wall_right {
			this.x = x_wall_right - palet_radius
			this.dir_x *= -1
		}
	}
}

func handlePaletCollision(p1, p2 *palet) {
	nx, ny, d := normalize(p2.x-p1.x, p2.y-p1.y)
	if d <= 2*palet_radius {
		tx := -ny
		ty := nx

		v1n := (p1.dir_x*nx + p1.dir_y*ny) * p1.v
		v1t := (p1.dir_x*tx + p1.dir_y*ty) * p1.v
		v2n := (p2.dir_x*nx + p2.dir_y*ny) * p2.v
		v2t := (p2.dir_x*tx + p2.dir_y*ty) * p2.v

		p1.dir_x = v2n*nx + v1t*tx
		p1.dir_y = v2n*ny + v1t*ty
		p2.dir_x = v1n*nx + v2t*tx
		p2.dir_y = v1n*ny + v2t*ty

		p1.dir_x, p1.dir_y, p1.v = normalize(p1.dir_x, p1.dir_y)
		p2.dir_x, p2.dir_y, p2.v = normalize(p2.dir_x, p2.dir_y)
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
			palets[k].x = width / 4.0 * float64(1+2*i)
			palets[k].y = height / 8.0 * float64(1+2*j)
			k++
		}
	}

	return PaletGame{&palets}
}

var accumulator time.Duration

func (this PaletGame) Step(dt time.Duration) {
	const acceleration = -10.0
	timestep, err := time.ParseDuration("1ms")
	fatal.Check(err)
	accumulator += dt

	for accumulator >= timestep {
		for i := range this.Palets {
			this.Palets[i].x += this.Palets[i].dir_x *
				this.Palets[i].v * timestep.Seconds()
			this.Palets[i].y += this.Palets[i].dir_y *
				this.Palets[i].v * timestep.Seconds()

			this.Palets[i].v += acceleration * timestep.Seconds()
			if this.Palets[i].v < 0.0 {
				this.Palets[i].v = 0.0
			}

			this.Palets[i].handleBoardCollision()

			for j := range this.Palets[i+1:] {
				p1 := &this.Palets[i]
				p2 := &this.Palets[i+j+1]
				handlePaletCollision(p1, p2)
			}
		}
		accumulator -= timestep
	}
}
