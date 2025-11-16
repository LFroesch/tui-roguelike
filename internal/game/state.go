package game

// GameState represents the current state of the game
type GameState struct {
	Player     *Entity    `json:"player"`
	Enemies    []*Entity  `json:"enemies"`
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

	// Multiplayer fields
	Players map[string]*Entity `json:"players,omitempty"` // For multiplayer
}

// NewGameState creates a new game state
func NewGameState(player *Entity, mapWidth, mapHeight int) *GameState {
	dungeonMap := make([][]rune, mapHeight)
	visible := make([][]bool, mapHeight)
	explored := make([][]bool, mapHeight)

	for y := 0; y < mapHeight; y++ {
		dungeonMap[y] = make([]rune, mapWidth)
		visible[y] = make([]bool, mapWidth)
		explored[y] = make([]bool, mapWidth)
	}

	return &GameState{
		Player:     player,
		Enemies:    []*Entity{},
		Items:      []Item{},
		ItemPos:    []Position{},
		DungeonMap: dungeonMap,
		Visible:    visible,
		Explored:   explored,
		Floor:      1,
		Turn:       1,
		Messages:   []string{"Welcome to the Mini Roguelike!"},
		GameOver:   false,
		Victory:    false,
		Players:    make(map[string]*Entity),
	}
}

// AddMessage adds a message to the log with word wrapping
func (gs *GameState) AddMessage(msg string, maxLength int) {
	if maxLength == 0 {
		maxLength = 40
	}

	if len(msg) <= maxLength {
		gs.Messages = append(gs.Messages, msg)
	} else {
		// Simple word wrapping
		words := splitWords(msg)
		var currentLine string

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if len(testLine) <= maxLength {
				currentLine = testLine
			} else {
				if currentLine != "" {
					gs.Messages = append(gs.Messages, currentLine)
				}
				currentLine = word
			}
		}

		if currentLine != "" {
			gs.Messages = append(gs.Messages, currentLine)
		}
	}

	// Keep only last 10 messages
	if len(gs.Messages) > 10 {
		gs.Messages = gs.Messages[len(gs.Messages)-10:]
	}
}

// ClearVisibility clears the visibility map
func (gs *GameState) ClearVisibility() {
	for y := range gs.Visible {
		for x := range gs.Visible[y] {
			gs.Visible[y][x] = false
		}
	}
}

// ResetExploration clears the explored map
func (gs *GameState) ResetExploration() {
	for y := range gs.Explored {
		for x := range gs.Explored[y] {
			gs.Explored[y][x] = false
		}
	}
}

// Helper function to split string into words
func splitWords(s string) []string {
	var words []string
	var current string

	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}
