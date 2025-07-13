# Mini Roguelike

A terminal-based roguelike game built with Go and the Bubble Tea TUI framework. Explore randomly generated dungeons, fight enemies, collect treasure, and see how deep you can go!

## Features

- **Procedural Dungeon Generation** - Each floor features randomly generated rooms connected by corridors
- **Turn-based Combat** - Fight goblins and skeletons with tactical positioning
- **Character Progression** - Level up to gain HP, attack, and defense bonuses
- **Inventory System** - Collect and equip weapons, armor, and consumables
- **Line of Sight** - Fog of war system with limited vision radius
- **Save/Load Game** - Continue your adventure across sessions
- **Colorized Messages** - Rich terminal colors for better gameplay feedback

## Screenshots

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                              MINI ROGUELIKE                                                 │
│━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│███████████████████████████████████████│📜 MESSAGES                               │🎒 INVENTORY             │
│█................█████████████████....█│━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│━━━━━━━━━━━━━━━━━━━━━━━━━│
│█................█████████████████....█│  • Welcome to the Mini Roguelike!         │  ┌─ Empty ─┐            │
│█................█████████████████....█│  • You hit the Goblin for 8 damage!       │  └─────────┘            │
│█................█████████████████....█│  • You defeated the Goblin!               │                         │
│█....@...........█████████████████....█│  • (+15 XP, +8 gold)                      │    STATS                │
│█................█████████████████....█│                                           │ ━━━━━━━━━━━━━━━━━━━━━━━━│
│█................█████████████████....█│                                           │  LV 1         GP 8      │
│█................███████..............█│                                           │  HP 20/20     DF 5      │
│█................███████..............█│                                           │  XP 15/100    SP 6      │
│█................███████..............█│                                           │  AT 10        LK 4      │
│███████████████████████████████████████│                                           │  FL B1                  │
│                                       │                                           │                         │
│━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│WASD/Arrows: Move • Space: Attack • I: Inventory • Enter: Stairs • Ctrl+S: Save • L: Load • Q: Quit          │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.19 or higher
- Terminal with color support

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd mini-roguelike

# Initialize Go module (if needed)
go mod init tui-roguelike

# Install dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles/key

# Build the game
go build -o roguelike main.go

# Run the game
./roguelike
```

## How to Play

### Controls

- **WASD** or **Arrow Keys** - Move your character (@)
- **Space** - Attack adjacent enemies
- **I** - Toggle inventory view (currently unused)
- **Enter** - Use stairs to descend to the next floor
- **Ctrl+S** - Save your current game
- **L** - Load a saved game
- **Q** - Quit the game

### Game Elements

| Symbol | Description |
|--------|-------------|
| `@` | Player character |
| `█` | Wall |
| `.` | Floor |
| `G` | Goblin enemy |
| `S` | Skeleton enemy |
| `T` | Treasure chest |
| `>` | Stairs to next floor |

### Combat

- Combat is turn-based - you move, then enemies move
- Attack enemies by moving into them
- Different enemies have different stats:
  - **Goblins** - Weaker but more common
  - **Skeletons** - Stronger with more HP and attack

### Items

Find treasure chests throughout the dungeon containing:

- **Health Potions** - Restore HP immediately
- **Gold Coins** - Currency (currently for scoring)
- **Iron Sword** - Increases attack power
- **Leather Armor** - Increases defense

### Progression

- Gain XP by defeating enemies
- Level up to increase HP, attack, and defense
- Each floor becomes progressively more challenging
- Your goal is to survive as long as possible and collect as much gold as you can

## Game Mechanics

### Vision System

- You can only see within a limited radius around your character
- Areas you've explored remain visible but dimmed
- Use line-of-sight mechanics for tactical gameplay

### Dungeon Generation

- Each floor is procedurally generated with 3-6 rooms
- Rooms are connected by L-shaped corridors
- Player always starts in the first room
- Stairs to the next level are in the last room

### Save System

- Games are automatically saved to `roguelike_save.json`
- Use Ctrl+S to save manually
- Use L to load your saved game

## Development

### Project Structure

```
mini-roguelike/
├── main.go                 # Main game file
├── roguelike_save.json     # Save file (generated)
```

### Key Components

- **GameState** - Manages all game data (player, enemies, map, etc.)
- **Model** - Bubble Tea model implementing the TUI
- **Dungeon Generation** - Procedural room and corridor creation
- **Combat System** - Turn-based battle mechanics
- **Visibility System** - Line-of-sight and fog of war


## License

This project is open source. Feel free to modify and distribute.

## Contributing

Contributions are welcome! Some ideas for improvements:

- More enemy types and behaviors
- Additional items and equipment
- Magic system
- Boss battles
- Multiple victory conditions
- Sound effects
- Configuration options
