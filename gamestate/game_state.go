package gamestate

import (
	"io"
	"math"
	"time"
)

type palet struct {
	x    float64
	y    float64
	v    float64
	dirX float64
	dirY float64
}

func normalize(x, y float64) (float64, float64, float64) {
	norm := math.Hypot(x, y)
	return x / norm, y / norm, norm
}

func (p *palet) launch(dirX, dirY int) {
	p.dirX, p.dirY, _ = normalize(float64(dirX)-p.x, float64(dirY)-p.y)
	const v = 300.0
	p.v = v
}

const paletRadius = 25.0
const (
	xLeft      = 19.0
	xRight     = 420.0
	yTop       = 19.0
	yBottom    = 600.0
	wallWidth  = 120
	wallHeight = 40
)

var (
	yMid        = ((yBottom - yTop + 1) / 2) + yTop
	yWallTop    = yMid - wallHeight/2
	yWallBottom = yMid + wallHeight/2 - 1
	xWallLeft   = xLeft + wallWidth
	xWallRight  = xRight - wallWidth
)

func (p *palet) handleBoardCollision() {
	if p.x-paletRadius <= xLeft {
		p.x = xLeft + paletRadius
		p.dirX *= -1
	} else if p.x+paletRadius >= xRight {
		p.x = xRight - paletRadius
		p.dirX *= -1
	}

	if p.y-paletRadius <= yTop {
		p.y = yTop + paletRadius
		p.dirY *= -1
	} else if p.y+paletRadius >= yBottom {
		p.y = yBottom - paletRadius
		p.dirY *= -1
	}

	if p.x < xWallLeft || p.x > xWallRight {
		if p.y < yMid {
			if p.y+paletRadius >= yWallTop {
				p.y = yWallTop - paletRadius
				p.dirY *= -1
			}
		} else {
			if p.y-paletRadius <= yWallBottom {
				p.y = yWallBottom + paletRadius
				p.dirY *= -1
			}
		}
	}

	nx, ny, d := normalize(p.x-(xWallLeft+1), p.y-yMid)
	if d <= wallHeight+paletRadius {
		tx := -ny
		ty := nx
		vn := (p.dirX*nx + p.dirY*ny) * p.v
		vt := (p.dirX*tx + p.dirY*ty) * p.v
		p.dirX = -vn*nx + vt*tx
		p.dirY = -vn*ny + vt*ty
		p.dirX, p.dirY, p.v = normalize(p.dirX, p.dirY)
	}
	nx, ny, d = normalize(p.x-(xWallRight-1), p.y-yMid)
	if d <= wallHeight+paletRadius {
		tx := -ny
		ty := nx
		vn := (p.dirX*nx + p.dirY*ny) * p.v
		vt := (p.dirX*tx + p.dirY*ty) * p.v
		p.dirX = -vn*nx + vt*tx
		p.dirY = -vn*ny + vt*ty
		p.dirX, p.dirY, p.v = normalize(p.dirX, p.dirY)
	}
}

func handlePaletCollision(p1, p2 *palet) {
	nx, ny, d := normalize(p2.x-p1.x, p2.y-p1.y)
	if d <= 2*paletRadius {
		tx := -ny
		ty := nx

		v1n := (p1.dirX*nx + p1.dirY*ny) * p1.v
		v1t := (p1.dirX*tx + p1.dirY*ty) * p1.v
		v2n := (p2.dirX*nx + p2.dirY*ny) * p2.v
		v2t := (p2.dirX*tx + p2.dirY*ty) * p2.v

		p1.dirX = v2n*nx + v1t*tx
		p1.dirY = v2n*ny + v1t*ty
		p2.dirX = v1n*nx + v2t*tx
		p2.dirY = v1n*ny + v2t*ty

		p1.dirX, p1.dirY, p1.v = normalize(p1.dirX, p1.dirY)
		p2.dirX, p2.dirY, p2.v = normalize(p2.dirX, p2.dirY)
	}
}

type GameState struct {
	accumulator time.Duration
	palets      [8]palet
}

func NewGameState() *GameState {
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

	return &GameState{palets: palets}
}

func (g *GameState) X(palet int) float64 {
	return g.palets[palet].x
}

func (g *GameState) Y(palet int) float64 {
	return g.palets[palet].y
}

func (g *GameState) Launch(palet int, dirX, dirY int) {
	g.palets[palet].launch(dirX, dirY)
}

func (g *GameState) Serialize(w io.Writer) {
}

func (g *GameState) Step(dt time.Duration) {
	timestep := time.Millisecond
	g.accumulator += dt

	for g.accumulator >= timestep {
		for i := range g.palets {
			g.palets[i].x += g.palets[i].dirX *
				g.palets[i].v * timestep.Seconds()
			g.palets[i].y += g.palets[i].dirY *
				g.palets[i].v * timestep.Seconds()

			const acceleration = -10.0
			g.palets[i].v += acceleration * timestep.Seconds()
			if g.palets[i].v < 0.0 {
				g.palets[i].v = 0.0
			}
		}

		for i := range g.palets {
			for j := range g.palets[i+1:] {
				p1 := &g.palets[i]
				p2 := &g.palets[i+j+1]
				handlePaletCollision(p1, p2)
			}
		}

		for i := range g.palets {
			g.palets[i].handleBoardCollision()

		}
		g.accumulator -= timestep
	}
}
