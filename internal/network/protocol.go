package network

import "tui-roguelike/internal/game"

// MessageType represents different message types
type MessageType string

const (
	// Client -> Server
	MsgTypeConnect    MessageType = "connect"
	MsgTypeMove       MessageType = "move"
	MsgTypeAttack     MessageType = "attack"
	MsgTypeChat       MessageType = "chat"
	MsgTypeDisconnect MessageType = "disconnect"
	MsgTypeUseItem    MessageType = "use_item"
	MsgTypePickup     MessageType = "pickup"

	// Server -> Client
	MsgTypeState      MessageType = "state"
	MsgTypeMessage    MessageType = "message"
	MsgTypePlayerJoin MessageType = "player_join"
	MsgTypePlayerLeft MessageType = "player_left"
	MsgTypeError      MessageType = "error"
)

// Message is the base message structure
type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

// ConnectPayload is sent when a player connects
type ConnectPayload struct {
	PlayerName string `json:"player_name"`
}

// MovePayload is sent when a player wants to move
type MovePayload struct {
	DX int `json:"dx"` // -1, 0, or 1
	DY int `json:"dy"` // -1, 0, or 1
}

// AttackPayload is sent when a player attacks
type AttackPayload struct {
	TargetID string `json:"target_id"`
}

// ChatPayload is sent for chat messages
type ChatPayload struct {
	Message string `json:"message"`
}

// StatePayload contains the game state update
type StatePayload struct {
	GameState *game.GameState `json:"game_state"`
	Tick      int             `json:"tick"`
}

// MessagePayload contains a message for the player
type MessagePayload struct {
	Text string `json:"text"`
}

// PlayerPayload contains player information
type PlayerPayload struct {
	Player *game.Entity `json:"player"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Error string `json:"error"`
}
