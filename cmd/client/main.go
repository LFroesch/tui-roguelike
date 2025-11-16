package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"tui-roguelike/internal/game"
	"tui-roguelike/internal/world"
	"tui-roguelike/pkg/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the Bubble Tea model
type Model struct {
	config     *config.Config
	state      *game.GameState
	quests     []game.Quest
	showInv    bool
	showQuests bool
	rng        *rand.Rand
	width      int
	height     int
	dungeon    *world.DungeonGenerator
	spawner    *world.Spawner
	visibility *world.VisibilitySystem
}

// Key bindings
var keyMap = struct {
	up, down, left, right, inventory, quests, stairs, save, quit key.Binding
}{
	up:        key.NewBinding(key.WithKeys("w", "W", "up"), key.WithHelp("w/↑", "up")),
	down:      key.NewBinding(key.WithKeys("s", "S", "down"), key.WithHelp("s/↓", "down")),
	left:      key.NewBinding(key.WithKeys("a", "A", "left"), key.WithHelp("a/←", "left")),
	right:     key.NewBinding(key.WithKeys("d", "D", "right"), key.WithHelp("d/→", "right")),
	inventory: key.NewBinding(key.WithKeys("i", "I"), key.WithHelp("i", "inventory")),
	quests:    key.NewBinding(key.WithKeys("q", "Q"), key.WithHelp("q", "quests")),
	stairs:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "stairs")),
	save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	quit:      key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
}

// Initialize the game
func initialModel(cfg *config.Config) Model {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	player := game.NewPlayer("Hero",
		game.Position{X: 5, Y: 5},
		cfg.Player.HP,
		cfg.Player.Attack,
		cfg.Player.Defense,
		cfg.Player.Speed,
		cfg.Player.Luck,
		cfg.Player.MaxXP,
	)

	state := game.NewGameState(player, cfg.Game.MapWidth, cfg.Game.MapHeight)

	m := Model{
		config:     cfg,
		state:      state,
		quests:     game.SimpleQuests(),
		rng:        rng,
		dungeon:    world.NewDungeonGenerator(cfg, rng),
		spawner:    world.NewSpawner(cfg, rng),
		visibility: world.NewVisibilitySystem(cfg),
	}

	m.generateLevel()
	return m
}

// generateLevel generates a new dungeon level using static maps
func (m *Model) generateLevel() {
	// Get the static map for current floor
	staticMap := world.GetStaticMap(m.state.Floor)

	// Load the map layout
	world.LoadStaticMap(m.state, staticMap)

	// Spawn enemies and items at predefined positions
	m.spawner.SpawnEnemiesAtPositions(m.state, staticMap.GetEnemySpawns())
	m.spawner.SpawnTreasureAtPositions(m.state, staticMap.GetItemSpawns())

	// Update visibility
	m.visibility.UpdateVisibility(m.state)
}

// movePlayer moves the player and handles interactions
func (m *Model) movePlayer(dx, dy int) {
	newX := m.state.Player.Pos.X + dx
	newY := m.state.Player.Pos.Y + dy

	// Check bounds
	if newX < 0 || newX >= m.config.Game.MapWidth || newY < 0 || newY >= m.config.Game.MapHeight {
		return
	}

	tile := m.state.DungeonMap[newY][newX]

	// Check for walls
	if tile == world.TileWall {
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
	case world.TileStairsDown:
		m.nextFloor()
	case world.TileTreasure:
		m.pickupTreasure(newX, newY)
	}

	m.state.Turn++
	m.visibility.UpdateVisibility(m.state)
	m.enemyTurn()
}

// attackEnemy handles player attacking an enemy
func (m *Model) attackEnemy(enemyIndex int) {
	enemy := m.state.Enemies[enemyIndex]

	damage := game.CalculateDamage(m.state.Player, enemy, m.rng)
	enemy.TakeDamage(damage)

	m.state.AddMessage(fmt.Sprintf("You hit %s for %d damage!", enemy.Name, damage), 40)

	if !enemy.Alive {
		xp, gold := m.spawner.GetReward(enemy.CreatureType)
		m.state.Player.XP += xp
		m.state.Player.Gold += gold

		m.state.AddMessage(fmt.Sprintf("Defeated %s! (+%d XP, +%d gold)", enemy.Name, xp, gold), 40)

		// Update quests
		m.quests = game.UpdateQuestProgress(m.quests, game.QuestTypeKill, enemy.CreatureType, 1)

		// Check for level up
		if m.state.Player.XP >= m.state.Player.MaxXP {
			m.levelUp()
		}
	}

	m.state.Turn++
	m.visibility.UpdateVisibility(m.state)
	m.enemyTurn()
}

// levelUp levels up the player
func (m *Model) levelUp() {
	cfg := m.config.Leveling
	m.state.Player.Level++
	m.state.Player.XP -= m.state.Player.MaxXP
	m.state.Player.MaxXP += cfg.XPPerLevel

	hpGain := m.rng.Intn(cfg.HPGainMax-cfg.HPGainMin+1) + cfg.HPGainMin
	m.state.Player.MaxHP += hpGain
	m.state.Player.HP = m.state.Player.MaxHP
	m.state.Player.Attack += cfg.AttackGain
	m.state.Player.Defense += cfg.DefenseGain

	m.state.AddMessage(fmt.Sprintf("Level up! Now level %d!", m.state.Player.Level), 40)
	m.state.AddMessage(fmt.Sprintf("+%d HP, +%d ATK, +%d DEF", hpGain, cfg.AttackGain, cfg.DefenseGain), 40)
}

// pickupTreasure handles picking up treasure
func (m *Model) pickupTreasure(x, y int) {
	for i, pos := range m.state.ItemPos {
		if pos.X == x && pos.Y == y && i < len(m.state.Items) {
			item := m.state.Items[i]

			switch item.Effect {
			case "heal":
				m.state.Player.Heal(item.Value)
				m.state.AddMessage(fmt.Sprintf("Drank %s, recovered %d HP!", item.Name, item.Value), 40)
			case "gold":
				m.state.Player.Gold += item.Value
				m.state.AddMessage(fmt.Sprintf("Found %d gold!", item.Value), 40)
				m.quests = game.UpdateQuestProgress(m.quests, game.QuestTypeCollect, "gold", item.Value)
			case "attack", "defense", "speed":
				if m.state.Player.EquipItem(&item) {
					m.state.AddMessage(fmt.Sprintf("Equipped %s! (+%d to stats)", item.Name, item.Value), 40)
				}
				m.state.Player.AddItem(item)
			}

			m.state.DungeonMap[y][x] = world.TileFloor
			m.state.Items = append(m.state.Items[:i], m.state.Items[i+1:]...)
			m.state.ItemPos = append(m.state.ItemPos[:i], m.state.ItemPos[i+1:]...)
			break
		}
	}
}

// enemyTurn handles all enemy turns
func (m *Model) enemyTurn() {
	for _, enemy := range m.state.Enemies {
		if !enemy.Alive {
			continue
		}

		px, py := m.state.Player.Pos.X, m.state.Player.Pos.Y
		ex, ey := enemy.Pos.X, enemy.Pos.Y

		dist := math.Sqrt(float64((px-ex)*(px-ex) + (py-ey)*(py-ey)))

		if dist <= 6 && world.HasLineOfSight(m.state, ex, ey, px, py) {
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

			// Check if attacking player
			if newX == px && newY == py {
				damage := game.CalculateDamage(enemy, m.state.Player, m.rng)
				m.state.Player.TakeDamage(damage)
				m.state.AddMessage(fmt.Sprintf("%s attacks for %d damage!", enemy.Name, damage), 40)

				if !m.state.Player.Alive {
					m.state.GameOver = true
					m.state.AddMessage("You died! Game Over.", 40)
				}
				continue
			}

			// Move enemy
			if newX >= 0 && newX < m.config.Game.MapWidth && newY >= 0 && newY < m.config.Game.MapHeight {
				if m.state.DungeonMap[newY][newX] != world.TileWall {
					enemy.Pos.X = newX
					enemy.Pos.Y = newY
				}
			}
		}
	}
}

// nextFloor goes to the next floor
func (m *Model) nextFloor() {
	m.state.Floor++
	m.state.AddMessage(fmt.Sprintf("Descended to floor B%d", m.state.Floor), 40)

	// Update quests
	m.quests = game.UpdateQuestProgress(m.quests, game.QuestTypeReach, "floor", m.state.Floor)

	m.state.ResetExploration()
	m.generateLevel()
}

// Bubble Tea interface
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
			if msg.String() == "r" || msg.String() == "R" {
				return initialModel(m.config), nil
			}
			if msg.String() == "esc" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
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
		case msg.String() == "Q" || msg.String() == "q":
			m.showQuests = !m.showQuests
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.state.GameOver {
		return renderGameOver(m.state)
	}

	if m.showQuests {
		return renderQuests(m.quests, m.state)
	}

	return renderGame(&m)
}

func main() {
	cfg, err := config.Load("configs/game.yaml")
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.Default()
	}

	p := tea.NewProgram(initialModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
