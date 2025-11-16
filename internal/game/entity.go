package game

import "math/rand"

// Position represents a 2D coordinate
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Item represents an item in the game
type Item struct {
	Name   string `json:"name"`
	Type   string `json:"type"` // weapon, armor, potion, gold, quest
	Value  int    `json:"value"`
	Count  int    `json:"count"`
	Effect string `json:"effect"` // heal, attack, defense, speed, etc
}

// Entity represents any game entity (player or creature)
type Entity struct {
	ID      string   `json:"id,omitempty"` // For multiplayer
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

	// Equipment slots
	Equipment Equipment `json:"equipment,omitempty"`

	// Quest tracking
	ActiveQuests []string `json:"active_quests,omitempty"`

	// Creature type (for enemies)
	CreatureType string `json:"creature_type,omitempty"`
}

// Equipment represents equipped items
type Equipment struct {
	Weapon   *Item `json:"weapon,omitempty"`
	Armor    *Item `json:"armor,omitempty"`
	Helmet   *Item `json:"helmet,omitempty"`
	Boots    *Item `json:"boots,omitempty"`
	Accessory *Item `json:"accessory,omitempty"`
}

// NewPlayer creates a new player entity
func NewPlayer(name string, startPos Position, hp, attack, defense, speed, luck, maxXP int) *Entity {
	return &Entity{
		Pos:     startPos,
		Symbol:  '@',
		Name:    name,
		HP:      hp,
		MaxHP:   hp,
		Attack:  attack,
		Defense: defense,
		Speed:   speed,
		Luck:    luck,
		XP:      0,
		MaxXP:   maxXP,
		Level:   1,
		Gold:    0,
		Alive:   true,
		Items:   []Item{},
		Equipment: Equipment{},
		ActiveQuests: []string{},
	}
}

// NewCreature creates a new creature entity
func NewCreature(creatureType string, symbol rune, name string, pos Position, hp, attack, defense, speed int) *Entity {
	return &Entity{
		Pos:          pos,
		Symbol:       symbol,
		Name:         name,
		HP:           hp,
		MaxHP:        hp,
		Attack:       attack,
		Defense:      defense,
		Speed:        speed,
		Luck:         0,
		Alive:        true,
		CreatureType: creatureType,
	}
}

// TakeDamage applies damage to the entity
func (e *Entity) TakeDamage(damage int) {
	if damage < 1 {
		damage = 1
	}
	e.HP -= damage
	if e.HP <= 0 {
		e.HP = 0
		e.Alive = false
	}
}

// Heal restores HP to the entity
func (e *Entity) Heal(amount int) {
	e.HP += amount
	if e.HP > e.MaxHP {
		e.HP = e.MaxHP
	}
}

// AddItem adds an item to the entity's inventory
func (e *Entity) AddItem(item Item) {
	e.Items = append(e.Items, item)
}

// HasItem checks if entity has an item by name
func (e *Entity) HasItem(itemName string) bool {
	for _, item := range e.Items {
		if item.Name == itemName {
			return true
		}
	}
	return false
}

// EquipItem equips an item to appropriate slot
func (e *Entity) EquipItem(item *Item) bool {
	switch item.Type {
	case "weapon":
		e.Equipment.Weapon = item
		return true
	case "armor":
		e.Equipment.Armor = item
		return true
	case "helmet":
		e.Equipment.Helmet = item
		return true
	case "boots":
		e.Equipment.Boots = item
		return true
	case "accessory":
		e.Equipment.Accessory = item
		return true
	}
	return false
}

// GetTotalAttack returns attack including equipment bonuses
func (e *Entity) GetTotalAttack() int {
	total := e.Attack
	if e.Equipment.Weapon != nil {
		total += e.Equipment.Weapon.Value
	}
	return total
}

// GetTotalDefense returns defense including equipment bonuses
func (e *Entity) GetTotalDefense() int {
	total := e.Defense
	if e.Equipment.Armor != nil {
		total += e.Equipment.Armor.Value
	}
	if e.Equipment.Helmet != nil {
		total += e.Equipment.Helmet.Value
	}
	return total
}

// CalculateDamage calculates damage with variance
func CalculateDamage(attacker, defender *Entity, rng *rand.Rand) int {
	baseDamage := attacker.GetTotalAttack() + rng.Intn(5)
	damage := baseDamage - defender.GetTotalDefense()
	if damage < 1 {
		damage = 1
	}
	return damage
}
