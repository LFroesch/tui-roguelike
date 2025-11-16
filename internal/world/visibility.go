package world

import (
	"math"

	"tui-roguelike/internal/game"
	"tui-roguelike/pkg/config"
)

// VisibilitySystem handles line of sight and fog of war
type VisibilitySystem struct {
	Config *config.Config
}

// NewVisibilitySystem creates a new visibility system
func NewVisibilitySystem(cfg *config.Config) *VisibilitySystem {
	return &VisibilitySystem{Config: cfg}
}

// UpdateVisibility updates what the player can see
func (vs *VisibilitySystem) UpdateVisibility(state *game.GameState) {
	state.ClearVisibility()

	px, py := state.Player.Pos.X, state.Player.Pos.Y
	radius := vs.Config.Game.SightRadius

	for y := 0; y < vs.Config.Game.MapHeight; y++ {
		for x := 0; x < vs.Config.Game.MapWidth; x++ {
			dist := math.Sqrt(float64((x-px)*(x-px) + (y-py)*(y-py)))
			if dist <= float64(radius) {
				if HasLineOfSight(state, px, py, x, y) {
					state.Visible[y][x] = true
					state.Explored[y][x] = true
				}
			}
		}
	}
}

// HasLineOfSight checks if there's a clear line of sight using Bresenham's algorithm
func HasLineOfSight(state *game.GameState, x0, y0, x1, y1 int) bool {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	sy := 1

	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		if x == x1 && y == y1 {
			return true
		}

		// Don't check starting position
		if x != x0 || y != y0 {
			if state.DungeonMap[y][x] == TileWall {
				return false
			}
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
