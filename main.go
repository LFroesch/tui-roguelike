package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"tui-roguelike/internal/logdog"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Game constants
const (
	MAP_WIDTH    = 35
	MAP_HEIGHT   = 12
	SIGHT_RADIUS = 5
	SAVE_FILE    = "roguelike_save.json"
)

// Tile types
const (
	WALL        = '█'
	FLOOR       = '.'
	DOOR        = '+'
	STAIRS_UP   = '<'
	STAIRS_DOWN = '>'
	TREASURE    = 'T'
	GOBLIN      = 'G'
	SKELETON    = 'S'
	PLAYER      = '@'
)

// Colors and styles
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ffff")).
			Bold(true)

	selectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9945ff")).
			Bold(true).
			Background(lipgloss.Color("#221133"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	playerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ff00")).
			Bold(true).
			Background(lipgloss.Color("#003300"))

	enemyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4444")).
			Bold(true).
			Background(lipgloss.Color("#330000"))

	treasureStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffff00")).
			Bold(true).
			Background(lipgloss.Color("#333300"))

	wallStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	floorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dddddd"))

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)
)

// Position represents a 2D coordinate
type Position struct {
	X, Y int
}

// Item represents an item in the game
type Item struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Value  int    `json:"value"`
	Count  int    `json:"count"`
	Effect string `json:"effect"`
}

// Entity represents any game entity
type Entity struct {
	Pos     Position `json:"pos"`
	Symbol  rune     `json:"symbol"`
	Name    string   `json:"name"`
	HP      int      `json:"hp"`
	MaxHP   int      `json:"max_hp"`
	Attack  int      `json:"attack"`
	Defense int      `json:"defense"`
	Speed   int      `json:"speed"`
	Luck    int      `json:"luck"`
	XP      int      `json:"xp"`
	MaxXP   int      `json:"max_xp"`
	Level   int      `json:"level"`
	Gold    int      `json:"gold"`
	Alive   bool     `json:"alive"`
	Items   []Item   `json:"items"`
}

// Room represents a dungeon room
type Room struct {
	X, Y, Width, Height int
}

// GameState represents the current state of the game
type GameState struct {
	Player     Entity     `json:"player"`
	Enemies    []Entity   `json:"enemies"`
	Items      []Item     `json:"items"`
	ItemPos    []Position `json:"item_pos"`
	DungeonMap [][]rune   `json:"dungeon_map"`
	Visible    [][]bool   `json:"visible"`
	Explored   [][]bool   `json:"explored"`
	Floor      int        `json:"floor"`
	Turn       int        `json:"turn"`
	Messages   []string   `json:"messages"`
	GameOver   bool       `json:"game_over"`
	Victory    bool       `json:"victory"`
}

// Model represents the Bubble Tea model
type Model struct {
	state   GameState
	showInv bool
	rand    *rand.Rand
	width   int
	height  int
}

// Key bindings
var keyMap = struct {
	up, down, left, right, attack, inventory, stairs, save, quit key.Binding
}{
	up:        key.NewBinding(key.WithKeys("w", "W", "up"), key.WithHelp("w/↑", "up")),
	down:      key.NewBinding(key.WithKeys("s", "S", "down"), key.WithHelp("s/↓", "down")),
	left:      key.NewBinding(key.WithKeys("a", "A", "left"), key.WithHelp("a/←", "left")),
	right:     key.NewBinding(key.WithKeys("d", "D", "right"), key.WithHelp("d/→", "right")),
	attack:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "attack")),
	inventory: key.NewBinding(key.WithKeys("i", "I"), key.WithHelp("i", "inventory")),
	stairs:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "stairs")),
	save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	quit:      key.NewBinding(key.WithKeys("q", "Q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// Initialize the game
func initialModel() Model {
	m := Model{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	m.newGame()
	return m
}

// Create a new game
func (m *Model) newGame() {
	m.state = GameState{
		Player: Entity{
			Pos:     Position{X: 5, Y: 5},
			Symbol:  PLAYER,
			Name:    "Player",
			HP:      20,
			MaxHP:   20,
			Attack:  10,
			Defense: 5,
			Speed:   6,
			Luck:    4,
			XP:      0,
			MaxXP:   100,
			Level:   1,
			Gold:    0,
			Alive:   true,
			Items:   []Item{},
		},
		Enemies:    []Entity{},
		Items:      []Item{},
		ItemPos:    []Position{},
		DungeonMap: make([][]rune, MAP_HEIGHT),
		Visible:    make([][]bool, MAP_HEIGHT),
		Explored:   make([][]bool, MAP_HEIGHT),
		Floor:      1,
		Turn:       1,
		Messages:   []string{"Welcome to the Mini Roguelike!"},
		GameOver:   false,
		Victory:    false,
	}

	// Initialize maps
	for y := 0; y < MAP_HEIGHT; y++ {
		m.state.DungeonMap[y] = make([]rune, MAP_WIDTH)
		m.state.Visible[y] = make([]bool, MAP_WIDTH)
		m.state.Explored[y] = make([]bool, MAP_WIDTH)
	}

	m.generateDungeon()
	m.updateVisibility()
}

// Generate a new dungeon level
func (m *Model) generateDungeon() {
	// Fill with walls
	for y := 0; y < MAP_HEIGHT; y++ {
		for x := 0; x < MAP_WIDTH; x++ {
			m.state.DungeonMap[y][x] = WALL
		}
	}

	// Generate rooms
	rooms := m.generateRooms()

	// Connect rooms with corridors
	for i := 0; i < len(rooms)-1; i++ {
		m.createCorridor(rooms[i], rooms[i+1])
	}

	// Place player in first room
	if len(rooms) > 0 {
		room := rooms[0]
		m.state.Player.Pos = Position{
			X: room.X + room.Width/2,
			Y: room.Y + room.Height/2,
		}
	}

	// Place stairs in last room
	if len(rooms) > 1 {
		room := rooms[len(rooms)-1]
		x := room.X + room.Width/2
		y := room.Y + room.Height/2
		m.state.DungeonMap[y][x] = STAIRS_DOWN
	}

	// Place enemies and treasure
	m.placeEnemies(rooms)
	m.placeTreasure(rooms)
}

// Generate rooms using simple algorithm
func (m *Model) generateRooms() []Room {
	var rooms []Room
	attempts := 0
	maxAttempts := 50

	for len(rooms) < 6 && attempts < maxAttempts {
		attempts++

		width := m.rand.Intn(6) + 4
		height := m.rand.Intn(4) + 3
		x := m.rand.Intn(MAP_WIDTH-width-2) + 1
		y := m.rand.Intn(MAP_HEIGHT-height-2) + 1

		newRoom := Room{X: x, Y: y, Width: width, Height: height}

		// Check if room overlaps with existing rooms
		overlaps := false
		for _, room := range rooms {
			if m.roomsOverlap(newRoom, room) {
				overlaps = true
				break
			}
		}

		if !overlaps {
			m.carveRoom(newRoom)
			rooms = append(rooms, newRoom)
		}
	}

	return rooms
}

// Check if two rooms overlap
func (m *Model) roomsOverlap(r1, r2 Room) bool {
	return r1.X < r2.X+r2.Width+1 && r1.X+r1.Width+1 > r2.X &&
		r1.Y < r2.Y+r2.Height+1 && r1.Y+r1.Height+1 > r2.Y
}

// Carve out a room
func (m *Model) carveRoom(room Room) {
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			m.state.DungeonMap[y][x] = FLOOR
		}
	}
}

// Create a corridor between two rooms
func (m *Model) createCorridor(r1, r2 Room) {
	x1 := r1.X + r1.Width/2
	y1 := r1.Y + r1.Height/2
	x2 := r2.X + r2.Width/2
	y2 := r2.Y + r2.Height/2

	// L-shaped corridor
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		m.state.DungeonMap[y1][x] = FLOOR
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		m.state.DungeonMap[y][x2] = FLOOR
	}
}

// Place enemies in rooms
func (m *Model) placeEnemies(rooms []Room) {
	m.state.Enemies = []Entity{}

	for i := 1; i < len(rooms); i++ { // Skip first room (player spawn)
		if m.rand.Float32() < 0.7 { // 70% chance of enemy in room
			room := rooms[i]
			enemy := Entity{
				Pos: Position{
					X: room.X + m.rand.Intn(room.Width),
					Y: room.Y + m.rand.Intn(room.Height),
				},
				Alive:   true,
				MaxHP:   m.rand.Intn(10) + 5,
				Attack:  m.rand.Intn(5) + 3,
				Defense: m.rand.Intn(3) + 1,
				Speed:   m.rand.Intn(3) + 3,
			}

			if m.rand.Float32() < 0.6 {
				enemy.Symbol = GOBLIN
				enemy.Name = "Goblin"
			} else {
				enemy.Symbol = SKELETON
				enemy.Name = "Skeleton"
				enemy.MaxHP += 3
				enemy.Attack += 2
			}

			enemy.HP = enemy.MaxHP
			m.state.Enemies = append(m.state.Enemies, enemy)
		}
	}
}

// Place treasure in rooms
func (m *Model) placeTreasure(rooms []Room) {
	m.state.Items = []Item{}
	m.state.ItemPos = []Position{}

	for i := 1; i < len(rooms); i++ {
		if m.rand.Float32() < 0.4 { // 40% chance of treasure
			room := rooms[i]
			pos := Position{
				X: room.X + m.rand.Intn(room.Width),
				Y: room.Y + m.rand.Intn(room.Height),
			}

			// Make sure position is empty
			empty := true
			for _, enemy := range m.state.Enemies {
				if enemy.Pos.X == pos.X && enemy.Pos.Y == pos.Y {
					empty = false
					break
				}
			}

			if empty {
				m.state.DungeonMap[pos.Y][pos.X] = TREASURE

				// Add random item
				items := []Item{
					{Name: "Health Potion", Type: "potion", Value: 10, Count: 1, Effect: "heal"},
					{Name: "Gold Coins", Type: "gold", Value: m.rand.Intn(20) + 10, Count: 1, Effect: "gold"},
					{Name: "Iron Sword", Type: "weapon", Value: 3, Count: 1, Effect: "attack"},
					{Name: "Leather Armor", Type: "armor", Value: 2, Count: 1, Effect: "defense"},
				}

				item := items[m.rand.Intn(len(items))]
				m.state.Items = append(m.state.Items, item)
				m.state.ItemPos = append(m.state.ItemPos, pos)
			}
		}
	}
}

// Update line of sight and visibility
func (m *Model) updateVisibility() {
	// Clear current visibility
	for y := 0; y < MAP_HEIGHT; y++ {
		for x := 0; x < MAP_WIDTH; x++ {
			m.state.Visible[y][x] = false
		}
	}

	// Calculate line of sight using simple circle
	px, py := m.state.Player.Pos.X, m.state.Player.Pos.Y

	for y := 0; y < MAP_HEIGHT; y++ {
		for x := 0; x < MAP_WIDTH; x++ {
			dist := math.Sqrt(float64((x-px)*(x-px) + (y-py)*(y-py)))
			if dist <= SIGHT_RADIUS {
				if m.hasLineOfSight(px, py, x, y) {
					m.state.Visible[y][x] = true
					m.state.Explored[y][x] = true
				}
			}
		}
	}
}

// Check line of sight using Bresenham's line algorithm
func (m *Model) hasLineOfSight(x0, y0, x1, y1 int) bool {
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

		if x != x0 || y != y0 { // Don't check starting position
			if m.state.DungeonMap[y][x] == WALL {
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

// Move player and handle interactions
func (m *Model) movePlayer(dx, dy int) {
	newX := m.state.Player.Pos.X + dx
	newY := m.state.Player.Pos.Y + dy

	// Check bounds
	if newX < 0 || newX >= MAP_WIDTH || newY < 0 || newY >= MAP_HEIGHT {
		return
	}

	tile := m.state.DungeonMap[newY][newX]

	// Check for walls
	if tile == WALL {
		return
	}

	// Check for enemies
	for i, enemy := range m.state.Enemies {
		if enemy.Pos.X == newX && enemy.Pos.Y == newY && enemy.Alive {
			m.attackEnemy(i)
			return
		}
	}

	// Move player
	m.state.Player.Pos.X = newX
	m.state.Player.Pos.Y = newY

	// Check for interactions
	switch tile {
	case STAIRS_DOWN:
		m.nextFloor()
	case TREASURE:
		m.pickupTreasure(newX, newY)
	}

	m.state.Turn++
	m.updateVisibility()

	m.enemyTurn()
}

// Attack an enemy
func (m *Model) attackEnemy(enemyIndex int) {
	enemy := &m.state.Enemies[enemyIndex]

	damage := m.state.Player.Attack + m.rand.Intn(5) - enemy.Defense
	if damage < 1 {
		damage = 1
	}

	enemy.HP -= damage
	m.addMessage(fmt.Sprintf("You hit the %s for %d damage!", enemy.Name, damage))

	if enemy.HP <= 0 {
		enemy.Alive = false
		xpGain := m.rand.Intn(20) + 10
		goldGain := m.rand.Intn(15) + 5

		m.state.Player.XP += xpGain
		m.state.Player.Gold += goldGain

		m.addMessage(fmt.Sprintf("You defeated the %s!", enemy.Name))
		m.addMessage(fmt.Sprintf("(+%d XP, +%d gold)", xpGain, goldGain))

		// Level up check
		if m.state.Player.XP >= m.state.Player.MaxXP {
			m.levelUp()
		}
	}

	m.state.Turn++
	m.updateVisibility()

	m.enemyTurn()
}

// Level up the player
func (m *Model) levelUp() {
	m.state.Player.Level++
	m.state.Player.XP -= m.state.Player.MaxXP
	m.state.Player.MaxXP += 50

	hpGain := m.rand.Intn(5) + 3
	m.state.Player.MaxHP += hpGain
	m.state.Player.HP = m.state.Player.MaxHP
	m.state.Player.Attack += 2
	m.state.Player.Defense += 1

	m.addMessage(fmt.Sprintf("Level up! You are now level %d!", m.state.Player.Level))
	m.addMessage(fmt.Sprintf("Gained: +%d HP, +2 ATK, +1 DEF", hpGain))
}

// Check if player already has an item by name
func (m *Model) hasItem(itemName string) bool {
	for _, item := range m.state.Player.Items {
		if item.Name == itemName {
			return true
		}
	}
	return false
}

// Pickup treasure
func (m *Model) pickupTreasure(x, y int) {
	for i, pos := range m.state.ItemPos {
		if pos.X == x && pos.Y == y && i < len(m.state.Items) {
			item := m.state.Items[i]

			switch item.Effect {
			case "heal":
				m.state.Player.HP = min(m.state.Player.HP+item.Value, m.state.Player.MaxHP)
				m.addMessage(fmt.Sprintf("You drink a %s and recover %d HP!", item.Name, item.Value))
			case "gold":
				m.state.Player.Gold += item.Value
				m.addMessage(fmt.Sprintf("You found %d gold pieces!", item.Value))
			case "attack":
				if m.hasItem(item.Name) {
					// Convert duplicate to coins
					m.state.Player.Gold += 20
					m.state.Player.Attack += item.Value
					m.addMessage(fmt.Sprintf("You already have %s! Sold for 20 gold. (+%d ATK)", item.Name, item.Value))
				} else {
					m.state.Player.Attack += item.Value
					m.addMessage(fmt.Sprintf("You equip %s! (+%d ATK)", item.Name, item.Value))
					m.state.Player.Items = append(m.state.Player.Items, item)
				}
			case "defense":
				if m.hasItem(item.Name) {
					// Convert duplicate to coins
					m.state.Player.Defense += item.Value
					m.state.Player.Gold += 20
					m.addMessage(fmt.Sprintf("You already have %s! Sold for 20 gold. (+%d DEF)", item.Name, item.Value))
				} else {
					m.state.Player.Defense += item.Value
					m.addMessage(fmt.Sprintf("You equip %s! (+%d DEF)", item.Name, item.Value))
					m.state.Player.Items = append(m.state.Player.Items, item)
				}
			}

			m.state.DungeonMap[y][x] = FLOOR
			// Remove item from lists
			m.state.Items = append(m.state.Items[:i], m.state.Items[i+1:]...)
			m.state.ItemPos = append(m.state.ItemPos[:i], m.state.ItemPos[i+1:]...)
			break
		}
	}
}

// Enemy turn
func (m *Model) enemyTurn() {
	for i := range m.state.Enemies {
		enemy := &m.state.Enemies[i]
		if !enemy.Alive {
			continue
		}

		// Simple AI: move towards player if in sight
		px, py := m.state.Player.Pos.X, m.state.Player.Pos.Y
		ex, ey := enemy.Pos.X, enemy.Pos.Y

		dist := math.Sqrt(float64((px-ex)*(px-ex) + (py-ey)*(py-ey)))

		if dist <= 6 && m.hasLineOfSight(ex, ey, px, py) {
			// Move towards player
			dx, dy := 0, 0
			if px > ex {
				dx = 1
			} else if px < ex {
				dx = -1
			}
			if py > ey {
				dy = 1
			} else if py < ey {
				dy = -1
			}

			newX, newY := ex+dx, ey+dy

			// Check if move is valid
			if newX >= 0 && newX < MAP_WIDTH && newY >= 0 && newY < MAP_HEIGHT {
				if m.state.DungeonMap[newY][newX] != WALL {
					// Check if attacking player
					if newX == px && newY == py {
						m.enemyAttackPlayer(i)
						continue
					}

					// Check if position is occupied by another enemy
					occupied := false
					for j, other := range m.state.Enemies {
						if j != i && other.Alive && other.Pos.X == newX && other.Pos.Y == newY {
							occupied = true
							break
						}
					}

					if !occupied {
						enemy.Pos.X = newX
						enemy.Pos.Y = newY
					}
				}
			}
		}
	}
}

// Enemy attacks player
func (m *Model) enemyAttackPlayer(enemyIndex int) {
	enemy := &m.state.Enemies[enemyIndex]

	damage := enemy.Attack + m.rand.Intn(3) - m.state.Player.Defense
	if damage < 1 {
		damage = 1
	}

	m.state.Player.HP -= damage
	m.addMessage(fmt.Sprintf("%s attacks you for %d damage!", enemy.Name, damage))

	if m.state.Player.HP <= 0 {
		m.state.GameOver = true
		m.addMessage("You have died! Game Over.")
	}
}

// Go to next floor
func (m *Model) nextFloor() {
	m.state.Floor++
	m.addMessage(fmt.Sprintf("You descend to floor B%d", m.state.Floor))

	// Reset visibility and exploration
	for y := 0; y < MAP_HEIGHT; y++ {
		for x := 0; x < MAP_WIDTH; x++ {
			m.state.Visible[y][x] = false
			m.state.Explored[y][x] = false
		}
	}

	m.generateDungeon()
	m.updateVisibility()
}

// Add message to log
func (m *Model) addMessage(msg string) {
	const MAX_MESSAGE_LENGTH = 40 // Approximate safe length for the message column

	// If message is short enough, add it directly
	if len(msg) <= MAX_MESSAGE_LENGTH {
		m.state.Messages = append(m.state.Messages, msg)
	} else {
		// Split long messages intelligently
		words := strings.Fields(msg)
		var currentLine string

		for _, word := range words {
			// Check if adding this word would make the line too long
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if len(testLine) <= MAX_MESSAGE_LENGTH {
				currentLine = testLine
			} else {
				// Current line is full, add it and start a new line
				if currentLine != "" {
					m.state.Messages = append(m.state.Messages, currentLine)
				}
				currentLine = word
			}
		}

		// Add any remaining text
		if currentLine != "" {
			m.state.Messages = append(m.state.Messages, currentLine)
		}
	}

	// Keep only the last 10 messages
	if len(m.state.Messages) > 10 {
		excess := len(m.state.Messages) - 10
		m.state.Messages = m.state.Messages[excess:]
	}
}

// Color code words in messages
func (m *Model) colorizeMessage(msg string) string {
	// Define color styles for different keywords with improved coordination
	goldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700")).Bold(true)     // true gold
	xpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9933ff")).Bold(true)       // purple for XP
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6600")).Bold(true)    // orange for stats
	youStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88")).Bold(true)      // bright green for player
	enemyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3333")).Bold(true)    // bright red for enemies
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ccff")).Bold(true)     // cyan for items
	numberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44aaff")).Bold(true)   // bright blue for numbers
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff44")).Bold(true)   // bright yellow for actions
	locationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cc66ff")).Bold(true) // light purple for locations
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44ff44")).Bold(true)  // success green
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Bold(true)    // error red

	// Split message into words and colorize keywords
	words := strings.Fields(msg)
	for i, word := range words {
		lower := strings.ToLower(word)

		// Remove punctuation for matching
		cleanWord := strings.Trim(lower, ".,!:;()[]")

		// Check if word contains numbers
		hasNumbers := false
		for _, char := range word {
			if char >= '0' && char <= '9' {
				hasNumbers = true
				break
			}
		}

		// Track if we found a match
		matched := false

		switch cleanWord {
		case "gold", "coins", "pieces", "treasure", "money", "gp", "sold":
			words[i] = goldStyle.Render(word)
			matched = true
		case "xp", "experience", "exp", "gained":
			words[i] = xpStyle.Render(word)
			matched = true
		case "hp", "atk", "def", "attack", "defense", "level", "damage", "health", "strength", "up":
			words[i] = statsStyle.Render(word)
			matched = true
		case "you", "your", "yourself", "player":
			words[i] = youStyle.Render(word)
			matched = true
		case "goblin", "skeleton", "enemy", "enemies", "monster", "monsters", "foe", "foes":
			words[i] = enemyStyle.Render(word)
			matched = true
		case "potion", "sword", "armor", "weapon", "equipment", "item", "items", "leather", "iron":
			words[i] = itemStyle.Render(word)
			matched = true
		case "hit", "hits", "attacks", "defeated", "killed", "dies", "died", "found", "picked", "equip", "drink", "use", "recover", "already", "have":
			words[i] = actionStyle.Render(word)
			matched = true
		case "floor", "dungeon", "stairs", "room", "descend", "ascend":
			words[i] = locationStyle.Render(word)
			matched = true
		case "successfully", "saved", "loaded", "welcome":
			words[i] = successStyle.Render(word)
			matched = true
		case "error", "failed", "over", "game":
			words[i] = errorStyle.Render(word)
			matched = true
		default:
			// Color numbers differently
			if hasNumbers && len(cleanWord) <= 4 {
				words[i] = numberStyle.Render(word)
				matched = true
			}
		}

		// If no match found, render with actionStyle (yellow)
		if !matched {
			words[i] = actionStyle.Render(word)
		}
	}

	return strings.Join(words, " ")
}

// Save game to file
func (m *Model) saveGame() {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		m.addMessage("Error saving game!")
		return
	}

	err = os.WriteFile(SAVE_FILE, data, 0644)
	if err != nil {
		m.addMessage("Error writing save file!")
		return
	}

	m.addMessage("Game saved successfully!")
}

// Load game from file
func (m *Model) loadGame() {
	data, err := os.ReadFile(SAVE_FILE)
	if err != nil {
		m.addMessage("No save file found!")
		return
	}

	err = json.Unmarshal(data, &m.state)
	if err != nil {
		m.addMessage("Error loading save file!")
		return
	}

	m.updateVisibility()
	m.addMessage("Game loaded successfully!")
}

// Utility functions
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

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// Bubble Tea interface methods
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.state.GameOver {
			if msg.String() == "q" || msg.String() == "Q" {
				return m, tea.Quit
			}
			if msg.String() == "r" || msg.String() == "R" {
				m.newGame()
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keyMap.quit):
			return m, tea.Quit
		case key.Matches(msg, keyMap.up):
			m.movePlayer(0, -1)
		case key.Matches(msg, keyMap.down):
			m.movePlayer(0, 1)
		case key.Matches(msg, keyMap.left):
			m.movePlayer(-1, 0)
		case key.Matches(msg, keyMap.right):
			m.movePlayer(1, 0)
		case key.Matches(msg, keyMap.inventory):
			m.showInv = !m.showInv
		case key.Matches(msg, keyMap.save):
			m.saveGame()
		case msg.String() == "l" || msg.String() == "L":
			m.loadGame()
		}
	}

	return m, nil
}

func (m Model) View() string {
	// Only show size warning if terminal is really too small to display the game
	const MIN_WIDTH = 125
	const MIN_HEIGHT = 16

	if m.width > 0 && m.height > 0 && (m.width < MIN_WIDTH || m.height < MIN_HEIGHT) {
		return borderStyle.Render(fmt.Sprintf(`
%s

Terminal size: %dx%d
Required: %dx%d minimum

Please resize your terminal and restart the game.

Press Q to quit.

		`, headerStyle.Render("Terminal Too Small"), m.width, m.height, MIN_WIDTH, MIN_HEIGHT))
	}

	if m.state.GameOver {
		return borderStyle.Render(fmt.Sprintf(`
%s

You died on floor B%d after %d turns.
You reached level %d and collected %d gold.

Press R to restart or Q to quit.

		`, headerStyle.Render("GAME OVER"), m.state.Floor, m.state.Turn, m.state.Player.Level, m.state.Player.Gold))
	}

	// Create the game map
	var mapLines []string
	for y := 0; y < MAP_HEIGHT; y++ {
		var line string
		for x := 0; x < MAP_WIDTH; x++ {
			char := ' '
			style := lipgloss.NewStyle()

			if m.state.Visible[y][x] {
				// Check for player
				if m.state.Player.Pos.X == x && m.state.Player.Pos.Y == y {
					char = PLAYER
					style = playerStyle
				} else {
					// Check for enemies
					enemyFound := false
					for _, enemy := range m.state.Enemies {
						if enemy.Pos.X == x && enemy.Pos.Y == y && enemy.Alive {
							char = enemy.Symbol
							style = enemyStyle
							enemyFound = true
							break
						}
					}

					if !enemyFound {
						char = m.state.DungeonMap[y][x]
						switch char {
						case WALL:
							style = wallStyle
						case FLOOR:
							style = floorStyle
						case TREASURE:
							style = treasureStyle
						case STAIRS_DOWN:
							style = selectionStyle
						}
					}
				}
			} else if m.state.Explored[y][x] {
				char = m.state.DungeonMap[y][x]
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
			}

			line += style.Render(string(char))
		}
		mapLines = append(mapLines, line)
	}

	// Column widths
	const MESSAGE_WIDTH = 45
	const SIDEBAR_WIDTH = 25

	// Create messages section (middle column)
	var messagesLines []string
	messagesLines = append(messagesLines, headerStyle.Render("📜 MESSAGES"))
	messagesLines = append(messagesLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("━", MESSAGE_WIDTH)))
	for _, msg := range m.state.Messages {
		colorizedMsg := m.colorizeMessage(msg)
		messagesLines = append(messagesLines, messageStyle.Render("  • "+colorizedMsg))
	}

	// Create right column (inventory + stats)
	var sidebarLines []string

	// Enhanced Inventory section
	sidebarLines = append(sidebarLines, headerStyle.Render("🎒 INVENTORY"))
	sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("━", SIDEBAR_WIDTH)))
	if len(m.state.Player.Items) == 0 {
		sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ┌─ Empty ─┐"))
		sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  └─────────┘"))
	} else {
		for _, item := range m.state.Player.Items {
			effect := ""
			if item.Effect == "attack" {
				effect = fmt.Sprintf("(+%d ATK)", item.Value)
			} else if item.Effect == "defense" {
				effect = fmt.Sprintf("(+%d DEF)", item.Value)
			}
			itemLine := fmt.Sprintf("%s %s", item.Name, effect)
			sidebarLines = append(sidebarLines, itemLine)
		}
	}

	sidebarLines = append(sidebarLines, "")
	sidebarLines = append(sidebarLines, headerStyle.Render("📊 STATS"))
	sidebarLines = append(sidebarLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("━", SIDEBAR_WIDTH)))

	// Enhanced stats display with bars and icons in 2 columns
	atkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6666")).Bold(true)
	defStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6666ff")).Bold(true)
	spdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff66")).Bold(true)
	lckStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#66ff66")).Bold(true)
	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0066")).Bold(true)
	xpBarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Bold(true)
	goldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00")).Bold(true)
	floorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9945ff")).Bold(true)
	levelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff")).Bold(true)

	// Create stats in 2 columns - left and right
	leftCol := []string{
		fmt.Sprintf("%s %d", levelStyle.Render("LV"), m.state.Player.Level),
		fmt.Sprintf("%s %d/%d", hpStyle.Render("HP"), m.state.Player.HP, m.state.Player.MaxHP),
		fmt.Sprintf("%s %d/%d", xpBarStyle.Render("XP"), m.state.Player.XP, m.state.Player.MaxXP),
		fmt.Sprintf("%s %d", atkStyle.Render("AT"), m.state.Player.Attack),
		fmt.Sprintf("%s B%d", floorStyle.Render("FL"), m.state.Floor),
	}

	rightCol := []string{
		fmt.Sprintf("%s %d", goldStyle.Render("GP"), m.state.Player.Gold),
		fmt.Sprintf("%s %d", defStyle.Render("DF"), m.state.Player.Defense),
		fmt.Sprintf("%s %d", spdStyle.Render("SP"), m.state.Player.Speed),
		fmt.Sprintf("%s %d", lckStyle.Render("LK"), m.state.Player.Luck),
		"", // Empty to balance columns
	}

	// Combine the columns with proper spacing
	for i := 0; i < len(leftCol); i++ {
		if leftCol[i] != "" || rightCol[i] != "" {
			line := fmt.Sprintf("  %-13s %s", leftCol[i], rightCol[i])
			sidebarLines = append(sidebarLines, line)
		}
	}

	// Pad all sections to match map height
	for len(messagesLines) < MAP_HEIGHT {
		messagesLines = append(messagesLines, "")
	}
	for len(sidebarLines) < MAP_HEIGHT {
		sidebarLines = append(sidebarLines, "")
	}

	// Create properly sized columns with fixed widths
	mapColumn := lipgloss.NewStyle().Width(MAP_WIDTH).Render(strings.Join(mapLines, "\n"))
	messagesColumn := lipgloss.NewStyle().Width(MESSAGE_WIDTH).Render(strings.Join(messagesLines, "\n"))
	sidebarColumn := lipgloss.NewStyle().Width(SIDEBAR_WIDTH).Render(strings.Join(sidebarLines, "\n"))

	// Create vertical separators that span the full height
	separatorLines := make([]string, MAP_HEIGHT)
	for i := 0; i < MAP_HEIGHT; i++ {
		separatorLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("│")
	}
	separator := strings.Join(separatorLines, "\n")

	// Join the three columns horizontally with separators
	gameContent := lipgloss.JoinHorizontal(lipgloss.Top, mapColumn, separator, messagesColumn, separator, sidebarColumn)

	// Create command bar with colored keys
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))     // Blue color for keys
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))  // Green color for action text
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Gray color for bullets

	commands := fmt.Sprintf("%s/%s: %s %s %s: %s %s %s: %s %s %s: %s %s %s: %s %s %s: %s %s %s: %s",
		keyStyle.Render("WASD"),
		keyStyle.Render("Arrows"),
		actionStyle.Render("Move"),
		bulletStyle.Render("•"),
		keyStyle.Render("Space"),
		actionStyle.Render("Attack"),
		bulletStyle.Render("•"),
		keyStyle.Render("I"),
		actionStyle.Render("Inventory"),
		bulletStyle.Render("•"),
		keyStyle.Render("Enter"),
		actionStyle.Render("Stairs"),
		bulletStyle.Render("•"),
		keyStyle.Render("Ctrl+S"),
		actionStyle.Render("Save"),
		bulletStyle.Render("•"),
		keyStyle.Render("L"),
		actionStyle.Render("Load"),
		bulletStyle.Render("•"),
		keyStyle.Render("Q"),
		actionStyle.Render("Quit"))

	// Assemble the full view - 3 column layout
	content := headerStyle.Render("MINI ROGUELIKE") + "\n"
	content += strings.Repeat("━", 110) + "\n"
	content += gameContent + "\n"
	content += strings.Repeat("━", 110) + "\n"
	content += commandStyle.Render(commands)

	return borderStyle.Render(content)
}

func main() {
	logdog.Info("Starting Mini Roguelike...")
	// Check if save file exists and offer to load
	if _, err := os.Stat(SAVE_FILE); err == nil {
		fmt.Println("Save file found! Loading existing game...")
	} else {
		fmt.Println("Starting new game...")
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
