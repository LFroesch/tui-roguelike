package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"tui-roguelike/internal/database"
	"tui-roguelike/internal/game"
	"tui-roguelike/internal/network"
	"tui-roguelike/internal/world"
	"tui-roguelike/pkg/config"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// GameServer manages the game state and clients
type GameServer struct {
	config     *config.Config
	db         *database.MongoDB
	state      *game.GameState
	clients    map[string]*Client
	clientsMux sync.RWMutex
	ticker     *time.Ticker
	tickCount  int
	rng        *rand.Rand
	dungeon    *world.DungeonGenerator
	spawner    *world.Spawner
	visibility *world.VisibilitySystem
}

// Client represents a connected player
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Player *game.Entity
	Send   chan network.Message
}

// NewGameServer creates a new game server
func NewGameServer(cfg *config.Config, db *database.MongoDB) *GameServer {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create a shared world for all players
	player := game.NewPlayer("Server", game.Position{X: 5, Y: 5},
		cfg.Player.HP, cfg.Player.Attack, cfg.Player.Defense,
		cfg.Player.Speed, cfg.Player.Luck, cfg.Player.MaxXP)

	state := game.NewGameState(player, cfg.Game.MapWidth, cfg.Game.MapHeight)

	server := &GameServer{
		config:     cfg,
		db:         db,
		state:      state,
		clients:    make(map[string]*Client),
		rng:        rng,
		dungeon:    world.NewDungeonGenerator(cfg, rng),
		spawner:    world.NewSpawner(cfg, rng),
		visibility: world.NewVisibilitySystem(cfg),
	}

	// Generate initial dungeon
	rooms := server.dungeon.Generate(server.state)
	server.spawner.SpawnEnemies(server.state, rooms)
	server.spawner.SpawnTreasure(server.state, rooms)

	return server
}

// Start starts the game server
func (gs *GameServer) Start() {
	gs.ticker = time.NewTicker(time.Duration(gs.config.Server.TickRate) * time.Millisecond)
	go gs.gameLoop()

	http.HandleFunc("/ws", gs.handleWebSocket)
	addr := fmt.Sprintf(":%d", gs.config.Server.Port)

	log.Printf("🎮 Game server starting on %s (Tick rate: %dms)", addr, gs.config.Server.TickRate)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// gameLoop is the main game tick loop
func (gs *GameServer) gameLoop() {
	for range gs.ticker.C {
		gs.tick()
	}
}

// tick processes one game tick
func (gs *GameServer) tick() {
	gs.clientsMux.RLock()
	defer gs.clientsMux.RUnlock()

	gs.tickCount++

	// Update enemies (simple AI)
	gs.updateEnemies()

	// Broadcast state to all clients
	gs.broadcastState()
}

// updateEnemies updates enemy AI and movement
func (gs *GameServer) updateEnemies() {
	// For now, enemies target the nearest player
	if len(gs.clients) == 0 {
		return
	}

	for _, enemy := range gs.state.Enemies {
		if !enemy.Alive {
			continue
		}

		// Find nearest player
		var nearestPlayer *game.Entity
		minDist := float64(999999)

		for _, client := range gs.clients {
			dist := distance(enemy.Pos, client.Player.Pos)
			if dist < minDist {
				minDist = dist
				nearestPlayer = client.Player
			}
		}

		if nearestPlayer == nil {
			continue
		}

		// Simple AI: move towards nearest player if in range
		if minDist <= 6 && world.HasLineOfSight(gs.state, enemy.Pos.X, enemy.Pos.Y, nearestPlayer.Pos.X, nearestPlayer.Pos.Y) {
			moveEnemyTowards(gs.state, enemy, nearestPlayer.Pos)
		}
	}
}

// handleWebSocket handles new WebSocket connections
func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:   generateID(),
		Conn: conn,
		Send: make(chan network.Message, 256),
	}

	gs.clientsMux.Lock()
	gs.clients[client.ID] = client
	gs.clientsMux.Unlock()

	log.Printf("✅ New client connected: %s", client.ID)

	go client.writePump()
	go gs.readPump(client)
}

// readPump handles incoming messages from a client
func (gs *GameServer) readPump(client *Client) {
	defer func() {
		gs.removeClient(client)
		client.Conn.Close()
	}()

	for {
		var msg network.Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		gs.handleMessage(client, msg)
	}
}

// writePump sends messages to the client
func (c *Client) writePump() {
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}

// handleMessage processes incoming messages from clients
func (gs *GameServer) handleMessage(client *Client, msg network.Message) {
	switch msg.Type {
	case network.MsgTypeConnect:
		gs.handleConnect(client, msg)
	case network.MsgTypeMove:
		gs.handleMove(client, msg)
	case network.MsgTypeChat:
		gs.handleChat(client, msg)
	}
}

// handleConnect handles player connection
func (gs *GameServer) handleConnect(client *Client, msg network.Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	playerName, ok := payload["player_name"].(string)
	if !ok {
		playerName = "Player"
	}

	// Create player entity
	startPos := game.Position{X: 5 + gs.rng.Intn(5), Y: 5 + gs.rng.Intn(5)}
	player := game.NewPlayer(playerName, startPos,
		gs.config.Player.HP, gs.config.Player.Attack, gs.config.Player.Defense,
		gs.config.Player.Speed, gs.config.Player.Luck, gs.config.Player.MaxXP)
	player.ID = client.ID

	client.Player = player
	gs.state.Players[client.ID] = player

	log.Printf("👤 Player '%s' (ID: %s) joined the game", playerName, client.ID)

	// Send initial state
	gs.sendState(client)

	// Notify other players
	gs.broadcastMessage(fmt.Sprintf("%s joined the game!", playerName), client.ID)
}

// handleMove handles player movement
func (gs *GameServer) handleMove(client *Client, msg network.Message) {
	if client.Player == nil {
		return
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	dx, _ := payload["dx"].(float64)
	dy, _ := payload["dy"].(float64)

	newX := client.Player.Pos.X + int(dx)
	newY := client.Player.Pos.Y + int(dy)

	// Check bounds
	if newX < 0 || newX >= gs.config.Game.MapWidth || newY < 0 || newY >= gs.config.Game.MapHeight {
		return
	}

	// Check for walls
	if gs.state.DungeonMap[newY][newX] == world.TileWall {
		return
	}

	// Move player
	client.Player.Pos.X = newX
	client.Player.Pos.Y = newY
}

// handleChat handles chat messages
func (gs *GameServer) handleChat(client *Client, msg network.Message) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	message, ok := payload["message"].(string)
	if !ok {
		return
	}

	chatMsg := fmt.Sprintf("[%s]: %s", client.Player.Name, message)
	gs.broadcastMessage(chatMsg, "")
}

// sendState sends game state to a specific client
func (gs *GameServer) sendState(client *Client) {
	msg := network.Message{
		Type: network.MsgTypeState,
		Payload: network.StatePayload{
			GameState: gs.state,
			Tick:      gs.tickCount,
		},
	}
	client.Send <- msg
}

// broadcastState sends game state to all clients
func (gs *GameServer) broadcastState() {
	for _, client := range gs.clients {
		if client.Player != nil {
			gs.sendState(client)
		}
	}
}

// broadcastMessage sends a message to all clients (except excluded)
func (gs *GameServer) broadcastMessage(text string, excludeID string) {
	msg := network.Message{
		Type: network.MsgTypeMessage,
		Payload: network.MessagePayload{
			Text: text,
		},
	}

	gs.clientsMux.RLock()
	defer gs.clientsMux.RUnlock()

	for id, client := range gs.clients {
		if id != excludeID {
			client.Send <- msg
		}
	}
}

// removeClient removes a disconnected client
func (gs *GameServer) removeClient(client *Client) {
	gs.clientsMux.Lock()
	defer gs.clientsMux.Unlock()

	if client.Player != nil {
		log.Printf("👋 Player '%s' (ID: %s) left the game", client.Player.Name, client.ID)
		delete(gs.state.Players, client.ID)
		gs.broadcastMessage(fmt.Sprintf("%s left the game", client.Player.Name), client.ID)
	}

	delete(gs.clients, client.ID)
	close(client.Send)
}

// Helper functions

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func distance(p1, p2 game.Position) float64 {
	dx := float64(p2.X - p1.X)
	dy := float64(p2.Y - p1.Y)
	return dx*dx + dy*dy // squared distance is fine for comparison
}

func moveEnemyTowards(state *game.GameState, enemy *game.Entity, target game.Position) {
	dx, dy := 0, 0

	if target.X > enemy.Pos.X {
		dx = 1
	} else if target.X < enemy.Pos.X {
		dx = -1
	}

	if target.Y > enemy.Pos.Y {
		dy = 1
	} else if target.Y < enemy.Pos.Y {
		dy = -1
	}

	newX := enemy.Pos.X + dx
	newY := enemy.Pos.Y + dy

	// Check if move is valid
	if newX >= 0 && newX < len(state.DungeonMap[0]) && newY >= 0 && newY < len(state.DungeonMap) {
		if state.DungeonMap[newY][newX] != world.TileWall {
			enemy.Pos.X = newX
			enemy.Pos.Y = newY
		}
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load("configs/game.yaml")
	if err != nil {
		log.Printf("⚠️  Warning: Could not load config file, using defaults: %v", err)
		cfg = config.Default()
	}

	// Initialize MongoDB (optional - can run without it)
	var db *database.MongoDB
	if cfg.Database.URI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db, err = database.NewMongoDB(&cfg.Database)
		if err != nil {
			log.Printf("⚠️  Warning: Could not connect to MongoDB: %v", err)
			log.Println("   Server will run without persistence")
		} else {
			log.Println("✅ Connected to MongoDB")
			defer db.Close(ctx)
		}
	}

	// Create and start server
	server := NewGameServer(cfg, db)
	server.Start()
}
