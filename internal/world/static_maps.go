package world

import (
	"strings"

	"tui-roguelike/internal/game"
)

// StaticMap represents a pre-designed map layout
type StaticMap struct {
	Layout      string           // ASCII map layout
	PlayerStart game.Position    // Where player spawns
	StairsPos   game.Position    // Stairs location
	EnemySpawns []game.Position  // Enemy spawn points
	ItemSpawns  []game.Position  // Item spawn points
	Width       int
	Height      int
}

// GetStaticMap returns a static map for the given floor
func GetStaticMap(floor int) *StaticMap {
	maps := []StaticMap{
		// Floor 1 - Simple connected rooms
		{
			Layout: `
███████████████████████████████████
█.................................█
█..┌────────┐....................█
█..│........│....................█
█..│...@....│....................█
█..│........│....................█
█..│...G....│....................█
█..└────┬───┘....................█
█.......│........................█
█.......│........................█
█.......│.......┌────────┐.......█
█.......│.......│........│.......█
█.......└───────┤...T....│.......█
█...............│........│.......█
█...............│........├───>...█
█...............└────────┘.......█
█.................................█
███████████████████████████████████`,
			PlayerStart: game.Position{X: 7, Y: 4},
			StairsPos:   game.Position{X: 30, Y: 14},
			EnemySpawns: []game.Position{
				{X: 7, Y: 6},
			},
			ItemSpawns: []game.Position{
				{X: 21, Y: 12},
			},
		},

		// Floor 2 - L-shaped corridor
		{
			Layout: `
███████████████████████████████████
█.................................█
█.@...............................█
█.................................█
█████████████████┐................█
█................│................█
█................│................█
█..G.............│................█
█................│................█
█████████████████┘................█
█.................................█
█................┌────────┐.......█
█................│........│.......█
█................│...T....│.......█
█................│...S....│.......█
█................│........├───>...█
█................└────────┘.......█
███████████████████████████████████`,
			PlayerStart: game.Position{X: 2, Y: 2},
			StairsPos:   game.Position{X: 30, Y: 15},
			EnemySpawns: []game.Position{
				{X: 4, Y: 7},
				{X: 21, Y: 14},
			},
			ItemSpawns: []game.Position{
				{X: 21, Y: 13},
			},
		},

		// Floor 3 - Three connected rooms
		{
			Layout: `
███████████████████████████████████
█.................................█
█.┌──────┐.....┌──────┐.....┌────█
█.│......│.....│......│.....│....█
█.│..@...├─────┤..G...├─────┤.T..█
█.│......│.....│......│.....│....█
█.└──────┘.....└──────┘.....└────█
█.................................█
█.................................█
█.┌──────┐.....┌──────┐.....┌────█
█.│......│.....│......│.....│....█
█.│..T...├─────┤..S...├─────┤.>..█
█.│......│.....│......│.....│....█
█.└──────┘.....└──────┘.....└────█
█.................................█
███████████████████████████████████`,
			PlayerStart: game.Position{X: 5, Y: 4},
			StairsPos:   game.Position{X: 30, Y: 11},
			EnemySpawns: []game.Position{
				{X: 16, Y: 4},
				{X: 16, Y: 11},
			},
			ItemSpawns: []game.Position{
				{X: 30, Y: 4},
				{X: 5, Y: 11},
			},
		},

		// Floor 4 - Cross layout
		{
			Layout: `
███████████████████████████████████
█.................................█
█.........┌──────┐................█
█.........│......│................█
█.........│..G...│................█
█.........│......│................█
█.........└───┬──┘................█
█.............│...................█
█──────┐......│......┌────────────█
█......│......@......│............█
█.T....├──────┼──────┤............█
█......│......│......│............█
█──────┘......│......└────────────█
█.............│...................█
█.........┌───┴──┐................█
█.........│......│................█
█.........│..S...├───>............█
█.........│......│................█
█.........└──────┘................█
███████████████████████████████████`,
			PlayerStart: game.Position{X: 14, Y: 9},
			StairsPos:   game.Position{X: 28, Y: 16},
			EnemySpawns: []game.Position{
				{X: 14, Y: 4},
				{X: 14, Y: 16},
			},
			ItemSpawns: []game.Position{
				{X: 3, Y: 10},
			},
		},

		// Floor 5 - Boss arena with pillars
		{
			Layout: `
███████████████████████████████████
█.................................█
█.................................█
█..┌─────────────────────────┐...█
█..│.........................│...█
█..│.....█....@....█.........│...█
█..│.........................│...█
█..│.........................│...█
█..│..G..................S...│...█
█..│.........................│...█
█..│.....█.........█.........│...█
█..│.........................│...█
█..│..........>..............│...█
█..│.........................│...█
█..└─────────────────────────┘...█
█.................................█
███████████████████████████████████`,
			PlayerStart: game.Position{X: 15, Y: 5},
			StairsPos:   game.Position{X: 14, Y: 12},
			EnemySpawns: []game.Position{
				{X: 6, Y: 8},
				{X: 25, Y: 8},
			},
			ItemSpawns: []game.Position{
				{X: 10, Y: 5},
				{X: 20, Y: 10},
			},
		},
	}

	// Cycle through maps based on floor
	index := (floor - 1) % len(maps)
	return &maps[index]
}

// LoadStaticMap loads a static map into the game state
func LoadStaticMap(state *game.GameState, staticMap *StaticMap) {
	lines := strings.Split(strings.TrimSpace(staticMap.Layout), "\n")

	// Load the map layout
	for y, line := range lines {
		if y >= len(state.DungeonMap) {
			break
		}
		for x, char := range line {
			if x >= len(state.DungeonMap[y]) {
				break
			}

			// Convert box drawing characters to game tiles
			switch char {
			case '█':
				state.DungeonMap[y][x] = TileWall
			case '.', ' ':
				state.DungeonMap[y][x] = TileFloor
			case '┌', '┐', '└', '┘', '─', '│', '┬', '┴', '├', '┤':
				state.DungeonMap[y][x] = TileWall
			case '@':
				state.DungeonMap[y][x] = TileFloor // Player position
			case '>':
				state.DungeonMap[y][x] = TileStairsDown
			case 'T':
				state.DungeonMap[y][x] = TileFloor // Will place treasure
			case 'G', 'S':
				state.DungeonMap[y][x] = TileFloor // Will place enemy
			default:
				state.DungeonMap[y][x] = TileFloor
			}
		}
	}

	// Set player position
	state.Player.Pos = staticMap.PlayerStart

	// Place stairs
	if staticMap.StairsPos.X > 0 && staticMap.StairsPos.Y > 0 {
		state.DungeonMap[staticMap.StairsPos.Y][staticMap.StairsPos.X] = TileStairsDown
	}
}

// GetEnemySpawns returns enemy spawn positions for the map
func (sm *StaticMap) GetEnemySpawns() []game.Position {
	return sm.EnemySpawns
}

// GetItemSpawns returns item spawn positions for the map
func (sm *StaticMap) GetItemSpawns() []game.Position {
	return sm.ItemSpawns
}
