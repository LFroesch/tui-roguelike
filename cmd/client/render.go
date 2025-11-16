package main

import (
	"fmt"
	"strings"

	"tui-roguelike/internal/game"
	"tui-roguelike/internal/world"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00ffff")).
			Bold(true)

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

	stairsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9945ff")).
			Bold(true)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dddddd"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1)

	questCompleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00ff00")).
				Bold(true)

	questActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ffff00"))
)

// renderGame renders the main game view
func renderGame(m *Model) string {
	mapWidth := m.config.Game.MapWidth
	mapHeight := m.config.Game.MapHeight

	// Render map
	var mapLines []string
	for y := 0; y < mapHeight; y++ {
		var line string
		for x := 0; x < mapWidth; x++ {
			char := ' '
			style := lipgloss.NewStyle()

			if m.state.Visible[y][x] {
				// Check for player
				if m.state.Player.Pos.X == x && m.state.Player.Pos.Y == y {
					char = '@'
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
						case world.TileWall:
							style = wallStyle
						case world.TileFloor:
							style = floorStyle
						case world.TileTreasure:
							style = treasureStyle
						case world.TileStairsDown:
							style = stairsStyle
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

	// Render messages
	var messagesLines []string
	messagesLines = append(messagesLines, headerStyle.Render("📜 MESSAGES"))
	messagesLines = append(messagesLines, strings.Repeat("━", 45))
	for _, msg := range m.state.Messages {
		messagesLines = append(messagesLines, messageStyle.Render("  • "+msg))
	}

	// Render stats and inventory
	var sidebarLines []string
	sidebarLines = append(sidebarLines, headerStyle.Render("⚔️  HERO"))
	sidebarLines = append(sidebarLines, strings.Repeat("━", 25))

	// Stats
	sidebarLines = append(sidebarLines, fmt.Sprintf("❤️  HP: %d/%d", m.state.Player.HP, m.state.Player.MaxHP))
	sidebarLines = append(sidebarLines, fmt.Sprintf("⭐ LV: %d  XP: %d/%d", m.state.Player.Level, m.state.Player.XP, m.state.Player.MaxXP))
	sidebarLines = append(sidebarLines, fmt.Sprintf("💰 Gold: %d", m.state.Player.Gold))
	sidebarLines = append(sidebarLines, fmt.Sprintf("🗡️  ATK: %d  DEF: %d", m.state.Player.GetTotalAttack(), m.state.Player.GetTotalDefense()))
	sidebarLines = append(sidebarLines, fmt.Sprintf("🏰 Floor: B%d", m.state.Floor))
	sidebarLines = append(sidebarLines, "")

	// Equipment
	sidebarLines = append(sidebarLines, headerStyle.Render("🎒 EQUIPMENT"))
	sidebarLines = append(sidebarLines, strings.Repeat("━", 25))
	if m.state.Player.Equipment.Weapon != nil {
		sidebarLines = append(sidebarLines, fmt.Sprintf("⚔️  %s", m.state.Player.Equipment.Weapon.Name))
	}
	if m.state.Player.Equipment.Armor != nil {
		sidebarLines = append(sidebarLines, fmt.Sprintf("🛡️  %s", m.state.Player.Equipment.Armor.Name))
	}
	if m.state.Player.Equipment.Helmet != nil {
		sidebarLines = append(sidebarLines, fmt.Sprintf("⛑️  %s", m.state.Player.Equipment.Helmet.Name))
	}
	if len(m.state.Player.Items) == 0 && m.state.Player.Equipment.Weapon == nil {
		sidebarLines = append(sidebarLines, "  Empty")
	}
	sidebarLines = append(sidebarLines, "")

	// Active quests summary
	activeQuests := game.GetActiveQuests(m.quests)
	sidebarLines = append(sidebarLines, headerStyle.Render(fmt.Sprintf("📋 QUESTS (%d)", len(activeQuests))))
	sidebarLines = append(sidebarLines, strings.Repeat("━", 25))
	for i, quest := range activeQuests {
		if i < 2 { // Show only first 2 quests
			progress := fmt.Sprintf("%d/%d", quest.Current, quest.Required)
			sidebarLines = append(sidebarLines, fmt.Sprintf("• %s: %s", quest.Title, progress))
		}
	}
	if len(activeQuests) > 2 {
		sidebarLines = append(sidebarLines, fmt.Sprintf("  ...and %d more", len(activeQuests)-2))
	}
	sidebarLines = append(sidebarLines, "")
	sidebarLines = append(sidebarLines, "Press Q to view all quests")

	// Pad sections
	for len(messagesLines) < mapHeight {
		messagesLines = append(messagesLines, "")
	}
	for len(sidebarLines) < mapHeight {
		sidebarLines = append(sidebarLines, "")
	}

	// Create columns
	mapColumn := lipgloss.NewStyle().Width(mapWidth).Render(strings.Join(mapLines, "\n"))
	messagesColumn := lipgloss.NewStyle().Width(45).Render(strings.Join(messagesLines, "\n"))
	sidebarColumn := lipgloss.NewStyle().Width(25).Render(strings.Join(sidebarLines, "\n"))

	// Create separators
	separatorLines := make([]string, mapHeight)
	for i := 0; i < mapHeight; i++ {
		separatorLines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("│")
	}
	separator := strings.Join(separatorLines, "\n")

	// Join columns
	gameContent := lipgloss.JoinHorizontal(lipgloss.Top, mapColumn, separator, messagesColumn, separator, sidebarColumn)

	// Commands
	commands := "WASD/Arrows: Move  •  Enter: Stairs  •  Q: Quests  •  ESC: Quit"

	// Assemble view
	content := headerStyle.Render("⚔️  MINI ROGUELIKE MMO ⚔️") + "\n"
	content += strings.Repeat("━", 110) + "\n"
	content += gameContent + "\n"
	content += strings.Repeat("━", 110) + "\n"
	content += commands

	return borderStyle.Render(content)
}

// renderGameOver renders the game over screen
func renderGameOver(state *game.GameState) string {
	return borderStyle.Render(fmt.Sprintf(`
%s

You died on floor B%d after %d turns.
You reached level %d and collected %d gold.

Press R to restart or ESC to quit.

	`, headerStyle.Render("💀 GAME OVER 💀"), state.Floor, state.Turn, state.Player.Level, state.Player.Gold))
}

// renderQuests renders the quest screen
func renderQuests(quests []game.Quest, state *game.GameState) string {
	var lines []string
	lines = append(lines, headerStyle.Render("📋 QUEST LOG"))
	lines = append(lines, "")

	active := game.GetActiveQuests(quests)
	completed := game.GetCompletedQuests(quests)

	if len(active) > 0 {
		lines = append(lines, questActiveStyle.Render("▶ ACTIVE QUESTS:"))
		lines = append(lines, "")
		for _, quest := range active {
			lines = append(lines, fmt.Sprintf("  %s", questActiveStyle.Render(quest.Title)))
			lines = append(lines, fmt.Sprintf("  %s", quest.Description))
			lines = append(lines, fmt.Sprintf("  Progress: %d/%d", quest.Current, quest.Required))
			lines = append(lines, fmt.Sprintf("  Reward: %d gold, %d XP", quest.Reward.Gold, quest.Reward.XP))
			lines = append(lines, "")
		}
	}

	if len(completed) > 0 {
		lines = append(lines, questCompleteStyle.Render("✓ COMPLETED QUESTS:"))
		lines = append(lines, "")
		for _, quest := range completed {
			lines = append(lines, fmt.Sprintf("  %s", questCompleteStyle.Render("✓ "+quest.Title)))
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Press Q to close")

	return borderStyle.Render(strings.Join(lines, "\n"))
}
