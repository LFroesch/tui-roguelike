package world

import (
	"math/rand"

	"tui-roguelike/internal/game"
	"tui-roguelike/pkg/config"
)

// Tile types
const (
	TileWall      = '█'
	TileFloor     = '.'
	TileDoor      = '+'
	TileStairsUp   = '<'
	TileStairsDown = '>'
	TileTreasure   = 'T'
)

// Room represents a dungeon room
type Room struct {
	X, Y, Width, Height int
}

// DungeonGenerator handles dungeon generation
type DungeonGenerator struct {
	Config *config.Config
	Rand   *rand.Rand
}

// NewDungeonGenerator creates a new dungeon generator
func NewDungeonGenerator(cfg *config.Config, rng *rand.Rand) *DungeonGenerator {
	return &DungeonGenerator{
		Config: cfg,
		Rand:   rng,
	}
}

// Generate creates a new dungeon level
func (dg *DungeonGenerator) Generate(state *game.GameState) []Room {
	mapWidth := dg.Config.Game.MapWidth
	mapHeight := dg.Config.Game.MapHeight

	// Fill with walls
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			state.DungeonMap[y][x] = TileWall
		}
	}

	// Generate rooms
	rooms := dg.generateRooms(state)

	// Connect rooms with corridors
	for i := 0; i < len(rooms)-1; i++ {
		dg.createCorridor(state, rooms[i], rooms[i+1])
	}

	// Place player in first room
	if len(rooms) > 0 {
		room := rooms[0]
		state.Player.Pos = game.Position{
			X: room.X + room.Width/2,
			Y: room.Y + room.Height/2,
		}
	}

	// Place stairs in last room
	if len(rooms) > 1 {
		room := rooms[len(rooms)-1]
		x := room.X + room.Width/2
		y := room.Y + room.Height/2
		state.DungeonMap[y][x] = TileStairsDown
	}

	return rooms
}

// generateRooms creates rooms using simple algorithm
func (dg *DungeonGenerator) generateRooms(state *game.GameState) []Room {
	var rooms []Room
	attempts := 0
	cfg := dg.Config.World

	for len(rooms) < cfg.MaxRooms && attempts < cfg.RoomAttempts {
		attempts++

		width := dg.Rand.Intn(cfg.MaxRoomWidth-cfg.MinRoomWidth+1) + cfg.MinRoomWidth
		height := dg.Rand.Intn(cfg.MaxRoomHeight-cfg.MinRoomHeight+1) + cfg.MinRoomHeight
		x := dg.Rand.Intn(dg.Config.Game.MapWidth-width-2) + 1
		y := dg.Rand.Intn(dg.Config.Game.MapHeight-height-2) + 1

		newRoom := Room{X: x, Y: y, Width: width, Height: height}

		// Check if room overlaps with existing rooms
		overlaps := false
		for _, room := range rooms {
			if roomsOverlap(newRoom, room) {
				overlaps = true
				break
			}
		}

		if !overlaps {
			carveRoom(state, newRoom)
			rooms = append(rooms, newRoom)
		}
	}

	return rooms
}

// roomsOverlap checks if two rooms overlap
func roomsOverlap(r1, r2 Room) bool {
	return r1.X < r2.X+r2.Width+1 && r1.X+r1.Width+1 > r2.X &&
		r1.Y < r2.Y+r2.Height+1 && r1.Y+r1.Height+1 > r2.Y
}

// carveRoom carves out a room in the dungeon
func carveRoom(state *game.GameState, room Room) {
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			state.DungeonMap[y][x] = TileFloor
		}
	}
}

// createCorridor creates an L-shaped corridor between two rooms
func (dg *DungeonGenerator) createCorridor(state *game.GameState, r1, r2 Room) {
	x1 := r1.X + r1.Width/2
	y1 := r1.Y + r1.Height/2
	x2 := r2.X + r2.Width/2
	y2 := r2.Y + r2.Height/2

	// Horizontal then vertical
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		state.DungeonMap[y1][x] = TileFloor
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		state.DungeonMap[y][x2] = TileFloor
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
