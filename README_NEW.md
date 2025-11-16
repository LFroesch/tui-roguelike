# 🎮 TUI Roguelike MMO

A modular, multiplayer-ready terminal roguelike game built with Go and the Charmbracelet TUI framework. Features clean architecture, quest system, equipment slots, and MongoDB persistence.

---

## 🚀 Quick Start

### Play Solo (Standalone Client)
```bash
# Build and run the standalone TUI client
go build -o client ./cmd/client
./client
```

### Run Multiplayer Server
```bash
# Build and run the game server
go build -o server ./cmd/server
./server
```

---

## ✨ What's New?

### 🏗️ **Modular Architecture**
The game has been refactored from a single 1,167-line file into clean, maintainable packages:

```
tui-roguelike/
├── cmd/
│   ├── client/          # 🎨 Standalone TUI client
│   │   ├── main.go
│   │   └── render.go    # UI rendering logic
│   └── server/          # 🌐 Multiplayer game server
│       └── main.go
├── internal/
│   ├── game/            # 🎯 Core game logic
│   │   ├── entity.go    # Player, Enemy, Item types
│   │   ├── state.go     # Game state management
│   │   └── quest.go     # Quest system
│   ├── world/           # 🗺️  World generation
│   │   ├── dungeon.go   # Procedural dungeon generation
│   │   ├── spawner.go   # Enemy & treasure spawning
│   │   └── visibility.go # Line-of-sight & fog of war
│   ├── network/         # 📡 Client-server protocol
│   │   └── protocol.go  # Message types for multiplayer
│   └── database/        # 💾 Persistence layer
│       └── mongodb.go   # MongoDB integration
├── pkg/
│   └── config/          # ⚙️  Configuration system
│       └── config.go
├── configs/
│   └── game.yaml        # 📝 Game configuration
├── main.go              # 🕹️  Original version (still works!)
└── README.md
```

### 🎯 **Quest System**
Three simple quest types to start:
- **Kill Quests**: "Defeat 5 goblins"
- **Collection Quests**: "Collect 100 gold"
- **Exploration Quests**: "Reach floor 5"

Press **Q** in-game to view your quest log!

### 🎒 **Equipment System**
Proper equipment slots with stat bonuses:
- **Weapon** → +Attack
- **Armor** → +Defense
- **Helmet** → +Defense
- **Boots** → +Speed
- **Accessory** → Special effects

Equipment is automatically equipped when picked up and bonuses apply immediately.

### ⚙️ **Configuration System**
All game parameters are now configurable via `configs/game.yaml`:
- Map dimensions
- Creature stats and spawn rates
- Player starting stats
- Leveling progression
- Server tick rate
- MongoDB connection

### 🌐 **Multiplayer-Ready**
- **Server tick system**: 10 ticks/second (configurable)
- **WebSocket-based communication**
- **Shared persistent world**
- **See other players in real-time**
- **Server-authoritative game state**

### 💾 **MongoDB Cloud Persistence**
- Save/load player data
- Store quest progress
- Persistent world state
- Player authentication (ready to add)

---

## 🎮 Controls

### Movement
- **WASD** or **Arrow Keys** → Move
- **Enter** → Use stairs to next floor

### Actions
- **Q** → Toggle quest log
- **I** → Toggle inventory
- **ESC** → Quit game

### Game Elements
```
@ = Player      █ = Wall        G = Goblin      > = Stairs
. = Floor       T = Treasure    S = Skeleton
```

---

## 🛠️ Configuration

Edit `configs/game.yaml` to customize:

```yaml
game:
  map_width: 35
  map_height: 12
  sight_radius: 5

player:
  hp: 20
  attack: 10
  defense: 5

creatures:
  goblin:
    hp_min: 5
    hp_max: 15
    spawn_weight: 0.6
  skeleton:
    hp_min: 8
    hp_max: 18
    spawn_weight: 0.4

server:
  port: 8080
  tick_rate: 100  # milliseconds (10 TPS)

database:
  uri: "mongodb+srv://your-cluster.mongodb.net"
  name: "roguelike"
```

---

## 🔧 Development

### Build Everything
```bash
# Client
go build -o client ./cmd/client

# Server
go build -o server ./cmd/server

# Original version (still works!)
go build -o roguelike main.go
```

### Run Tests
```bash
go test ./...
```

### Add Dependencies
```bash
go get <package>
go mod tidy
```

---

## 🌟 Adding Features

### Add a New Creature Type

1. Edit `configs/game.yaml`:
```yaml
creatures:
  orc:
    symbol: 'O'
    name: "Orc"
    hp_min: 15
    hp_max: 25
    attack_min: 8
    attack_max: 12
    spawn_weight: 0.3
```

2. That's it! The spawner will automatically use it.

### Add a New Quest

Edit `internal/game/quest.go`:
```go
{
    ID:          "kill_skeletons_10",
    Title:       "Skeleton Crusher",
    Description: "Defeat 10 skeletons",
    Type:        QuestTypeKill,
    Target:      "skeleton",
    Required:    10,
    Reward: QuestReward{
        Gold: 150,
        XP:   300,
    },
}
```

### Add a New Item Type

Edit `internal/world/spawner.go`:
```go
items := []game.Item{
    {Name: "Magic Ring", Type: "accessory", Value: 5, Effect: "luck"},
    // ... existing items
}
```

---

## 🗄️ MongoDB Setup

### MongoDB Atlas (Cloud - Free Tier)
1. Create account at [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
2. Create a free cluster
3. Get connection string
4. Update `configs/game.yaml`:
```yaml
database:
  uri: "mongodb+srv://username:password@cluster.mongodb.net"
  name: "roguelike"
```

### Local MongoDB
```bash
# Run MongoDB in Docker
docker run -d -p 27017:27017 mongo

# Update config
database:
  uri: "mongodb://localhost:27017"
  name: "roguelike"
```

The server gracefully handles MongoDB being unavailable and runs without persistence.

---

## 📡 Multiplayer Architecture

### How It Works

1. **Server** runs a game loop at 10 TPS (ticks per second)
2. **Clients** connect via WebSocket
3. **Players** send input commands (move, attack, chat)
4. **Server** processes all inputs, updates game state
5. **Broadcast** updated state to all clients every tick
6. **Clients** render the state in their TUI

### Message Types
```go
// Client → Server
MsgTypeConnect    // Join game
MsgTypeMove       // Move player
MsgTypeAttack     // Attack enemy
MsgTypeChat       // Send chat message

// Server → Client
MsgTypeState      // Full game state update
MsgTypeMessage    // Game message/chat
MsgTypePlayerJoin // New player joined
MsgTypePlayerLeft // Player disconnected
```

### Server Endpoint
```
ws://localhost:8080/ws
```

---

## 🎯 Roadmap

### Phase 1: Core Systems ✅
- [x] Modular architecture
- [x] Configuration system
- [x] Quest system (kill, collect, reach)
- [x] Equipment slots
- [x] MongoDB integration

### Phase 2: Multiplayer (In Progress)
- [x] Server tick system
- [x] WebSocket protocol
- [ ] Client-server synchronization
- [ ] Multiple players in same world
- [ ] Chat system

### Phase 3: Content
- [ ] 10+ unique creatures
- [ ] 20+ quests
- [ ] Rare/epic equipment tiers
- [ ] Boss fights
- [ ] Special abilities/spells

### Phase 4: Polish
- [ ] Player accounts & authentication
- [ ] Leaderboards
- [ ] Achievements
- [ ] Sound effects (terminal bell!)
- [ ] ASCII art animations

---

## 🤝 Contributing

### Code Style
- Use `gofmt` for formatting
- Follow Go best practices
- Write tests for new features
- Update this README for major changes

### Testing
```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/game

# With coverage
go test -cover ./...
```

---

## 📝 License

MIT License - Feel free to use this for learning or building your own roguelike!

---

## 🙏 Credits

- **Charmbracelet** - Amazing TUI framework
- **MongoDB** - Persistence layer
- **Gorilla WebSocket** - Real-time communication

---

## 🐛 Known Issues

1. **MongoDB dependency**: Requires network access to download. If build fails, check your internet connection.
2. **Terminal size**: Requires at least 125x16 characters
3. **Multiplayer**: Currently server-only, client-to-server not fully integrated yet

---

## 💡 Tips

- **Start simple**: Run the standalone client first to learn the game
- **Config-driven**: Most balance changes can be made in `game.yaml`
- **Modular design**: Easy to swap out MongoDB for PostgreSQL/SQLite
- **Add creatures**: Just edit YAML, no code needed!

Happy dungeon crawling! ⚔️
