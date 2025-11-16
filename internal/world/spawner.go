package world

import (
	"math/rand"

	"tui-roguelike/internal/game"
	"tui-roguelike/pkg/config"
)

// Spawner handles entity and item spawning
type Spawner struct {
	Config *config.Config
	Rand   *rand.Rand
}

// NewSpawner creates a new spawner
func NewSpawner(cfg *config.Config, rng *rand.Rand) *Spawner {
	return &Spawner{
		Config: cfg,
		Rand:   rng,
	}
}

// SpawnEnemies spawns enemies in rooms
func (s *Spawner) SpawnEnemies(state *game.GameState, rooms []Room) {
	state.Enemies = []*game.Entity{}

	// Skip first room (player spawn)
	for i := 1; i < len(rooms); i++ {
		if s.Rand.Float32() < s.Config.World.EnemySpawnChance {
			room := rooms[i]
			enemy := s.createRandomEnemy(game.Position{
				X: room.X + s.Rand.Intn(room.Width),
				Y: room.Y + s.Rand.Intn(room.Height),
			})
			state.Enemies = append(state.Enemies, enemy)
		}
	}
}

// createRandomEnemy creates a random enemy based on config
func (s *Spawner) createRandomEnemy(pos game.Position) *game.Entity {
	// Select random creature based on spawn weights
	var totalWeight float32
	for _, cfg := range s.Config.Creatures {
		totalWeight += cfg.SpawnWeight
	}

	roll := s.Rand.Float32() * totalWeight
	var currentWeight float32

	for creatureType, cfg := range s.Config.Creatures {
		currentWeight += cfg.SpawnWeight
		if roll <= currentWeight {
			return s.createCreature(creatureType, cfg, pos)
		}
	}

	// Fallback to goblin
	return s.createCreature("goblin", s.Config.Creatures["goblin"], pos)
}

// createCreature creates a specific creature
func (s *Spawner) createCreature(creatureType string, cfg config.CreatureConfig, pos game.Position) *game.Entity {
	hp := s.Rand.Intn(cfg.HPMax-cfg.HPMin+1) + cfg.HPMin
	attack := s.Rand.Intn(cfg.AttackMax-cfg.AttackMin+1) + cfg.AttackMin
	defense := s.Rand.Intn(cfg.DefenseMax-cfg.DefenseMin+1) + cfg.DefenseMin
	speed := s.Rand.Intn(cfg.SpeedMax-cfg.SpeedMin+1) + cfg.SpeedMin

	return game.NewCreature(creatureType, cfg.Symbol, cfg.Name, pos, hp, attack, defense, speed)
}

// SpawnTreasure spawns treasure in rooms
func (s *Spawner) SpawnTreasure(state *game.GameState, rooms []Room) {
	state.Items = []game.Item{}
	state.ItemPos = []game.Position{}

	for i := 1; i < len(rooms); i++ {
		if s.Rand.Float32() < s.Config.World.TreasureSpawnChance {
			room := rooms[i]
			pos := game.Position{
				X: room.X + s.Rand.Intn(room.Width),
				Y: room.Y + s.Rand.Intn(room.Height),
			}

			// Make sure position is empty
			empty := true
			for _, enemy := range state.Enemies {
				if enemy.Pos.X == pos.X && enemy.Pos.Y == pos.Y {
					empty = false
					break
				}
			}

			if empty {
				state.DungeonMap[pos.Y][pos.X] = TileTreasure

				// Add random item
				item := s.randomItem()
				state.Items = append(state.Items, item)
				state.ItemPos = append(state.ItemPos, pos)
			}
		}
	}
}

// randomItem creates a random item
func (s *Spawner) randomItem() game.Item {
	items := []game.Item{
		{Name: "Health Potion", Type: "potion", Value: 10, Count: 1, Effect: "heal"},
		{Name: "Gold Coins", Type: "gold", Value: s.Rand.Intn(20) + 10, Count: 1, Effect: "gold"},
		{Name: "Iron Sword", Type: "weapon", Value: 3, Count: 1, Effect: "attack"},
		{Name: "Leather Armor", Type: "armor", Value: 2, Count: 1, Effect: "defense"},
		{Name: "Steel Helmet", Type: "helmet", Value: 1, Count: 1, Effect: "defense"},
		{Name: "Leather Boots", Type: "boots", Value: 1, Count: 1, Effect: "speed"},
	}

	return items[s.Rand.Intn(len(items))]
}

// SpawnEnemiesAtPositions spawns enemies at specific positions (for static maps)
func (s *Spawner) SpawnEnemiesAtPositions(state *game.GameState, positions []game.Position) {
	state.Enemies = []*game.Entity{}

	for _, pos := range positions {
		enemy := s.createRandomEnemy(pos)
		state.Enemies = append(state.Enemies, enemy)
	}
}

// SpawnTreasureAtPositions spawns treasure at specific positions (for static maps)
func (s *Spawner) SpawnTreasureAtPositions(state *game.GameState, positions []game.Position) {
	state.Items = []game.Item{}
	state.ItemPos = []game.Position{}

	for _, pos := range positions {
		state.DungeonMap[pos.Y][pos.X] = TileTreasure

		item := s.randomItem()
		state.Items = append(state.Items, item)
		state.ItemPos = append(state.ItemPos, pos)
	}
}

// GetReward calculates enemy kill rewards
func (s *Spawner) GetReward(creatureType string) (xp int, gold int) {
	cfg, ok := s.Config.Creatures[creatureType]
	if !ok {
		return 15, 10 // Default
	}

	xp = s.Rand.Intn(cfg.XPRewardMax-cfg.XPRewardMin+1) + cfg.XPRewardMin
	gold = s.Rand.Intn(cfg.GoldRewardMax-cfg.GoldRewardMin+1) + cfg.GoldRewardMin
	return
}
