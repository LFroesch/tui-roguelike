package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all game configuration
type Config struct {
	Game     GameConfig                `yaml:"game"`
	World    WorldConfig               `yaml:"world"`
	Player   PlayerConfig              `yaml:"player"`
	Leveling LevelingConfig            `yaml:"leveling"`
	Creatures map[string]CreatureConfig `yaml:"creatures"`
	Server   ServerConfig              `yaml:"server"`
	Database DatabaseConfig            `yaml:"database"`
}

type GameConfig struct {
	MapWidth    int    `yaml:"map_width"`
	MapHeight   int    `yaml:"map_height"`
	SightRadius int    `yaml:"sight_radius"`
	SaveFile    string `yaml:"save_file"`
}

type WorldConfig struct {
	MaxRooms           int     `yaml:"max_rooms"`
	MinRoomWidth       int     `yaml:"min_room_width"`
	MaxRoomWidth       int     `yaml:"max_room_width"`
	MinRoomHeight      int     `yaml:"min_room_height"`
	MaxRoomHeight      int     `yaml:"max_room_height"`
	RoomAttempts       int     `yaml:"room_attempts"`
	EnemySpawnChance   float32 `yaml:"enemy_spawn_chance"`
	TreasureSpawnChance float32 `yaml:"treasure_spawn_chance"`
}

type PlayerConfig struct {
	HP      int `yaml:"hp"`
	Attack  int `yaml:"attack"`
	Defense int `yaml:"defense"`
	Speed   int `yaml:"speed"`
	Luck    int `yaml:"luck"`
	MaxXP   int `yaml:"max_xp"`
}

type LevelingConfig struct {
	XPPerLevel   int `yaml:"xp_per_level"`
	HPGainMin    int `yaml:"hp_gain_min"`
	HPGainMax    int `yaml:"hp_gain_max"`
	AttackGain   int `yaml:"attack_gain"`
	DefenseGain  int `yaml:"defense_gain"`
}

type CreatureConfig struct {
	Symbol        rune    `yaml:"symbol"`
	Name          string  `yaml:"name"`
	HPMin         int     `yaml:"hp_min"`
	HPMax         int     `yaml:"hp_max"`
	AttackMin     int     `yaml:"attack_min"`
	AttackMax     int     `yaml:"attack_max"`
	DefenseMin    int     `yaml:"defense_min"`
	DefenseMax    int     `yaml:"defense_max"`
	SpeedMin      int     `yaml:"speed_min"`
	SpeedMax      int     `yaml:"speed_max"`
	XPRewardMin   int     `yaml:"xp_reward_min"`
	XPRewardMax   int     `yaml:"xp_reward_max"`
	GoldRewardMin int     `yaml:"gold_reward_min"`
	GoldRewardMax int     `yaml:"gold_reward_max"`
	SpawnWeight   float32 `yaml:"spawn_weight"`
}

type ServerConfig struct {
	Port       int `yaml:"port"`
	TickRate   int `yaml:"tick_rate"`
	MaxPlayers int `yaml:"max_players"`
}

type DatabaseConfig struct {
	URI     string `yaml:"uri"`
	Name    string `yaml:"name"`
	Timeout int    `yaml:"timeout"`
}

// Load reads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Game: GameConfig{
			MapWidth:    35,
			MapHeight:   12,
			SightRadius: 5,
			SaveFile:    "roguelike_save.json",
		},
		Player: PlayerConfig{
			HP:      20,
			Attack:  10,
			Defense: 5,
			Speed:   6,
			Luck:    4,
			MaxXP:   100,
		},
		Server: ServerConfig{
			Port:       8080,
			TickRate:   100,
			MaxPlayers: 100,
		},
	}
}
