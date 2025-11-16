package database

import (
	"context"
	"time"

	"tui-roguelike/internal/game"
	"tui-roguelike/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB wraps MongoDB client
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.DatabaseConfig
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.Name),
		config:   cfg,
	}, nil
}

// Close closes the MongoDB connection
func (db *MongoDB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

// SavePlayer saves or updates a player to the database
func (db *MongoDB) SavePlayer(ctx context.Context, player *game.Entity) error {
	collection := db.database.Collection("players")

	filter := bson.M{"id": player.ID}
	update := bson.M{"$set": player}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// LoadPlayer loads a player from the database
func (db *MongoDB) LoadPlayer(ctx context.Context, playerID string) (*game.Entity, error) {
	collection := db.database.Collection("players")

	var player game.Entity
	err := collection.FindOne(ctx, bson.M{"id": playerID}).Decode(&player)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Player not found
		}
		return nil, err
	}

	return &player, nil
}

// SaveGameState saves the entire game state (for single-player saves)
func (db *MongoDB) SaveGameState(ctx context.Context, playerID string, state *game.GameState) error {
	collection := db.database.Collection("game_states")

	filter := bson.M{"player_id": playerID}
	update := bson.M{
		"$set": bson.M{
			"player_id":  playerID,
			"state":      state,
			"updated_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// LoadGameState loads a saved game state
func (db *MongoDB) LoadGameState(ctx context.Context, playerID string) (*game.GameState, error) {
	collection := db.database.Collection("game_states")

	var result struct {
		PlayerID  string           `bson:"player_id"`
		State     *game.GameState  `bson:"state"`
		UpdatedAt time.Time        `bson:"updated_at"`
	}

	err := collection.FindOne(ctx, bson.M{"player_id": playerID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return result.State, nil
}

// SaveQuests saves player quests
func (db *MongoDB) SaveQuests(ctx context.Context, playerID string, quests []game.Quest) error {
	collection := db.database.Collection("player_quests")

	filter := bson.M{"player_id": playerID}
	update := bson.M{
		"$set": bson.M{
			"player_id":  playerID,
			"quests":     quests,
			"updated_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// LoadQuests loads player quests
func (db *MongoDB) LoadQuests(ctx context.Context, playerID string) ([]game.Quest, error) {
	collection := db.database.Collection("player_quests")

	var result struct {
		PlayerID  string       `bson:"player_id"`
		Quests    []game.Quest `bson:"quests"`
		UpdatedAt time.Time    `bson:"updated_at"`
	}

	err := collection.FindOne(ctx, bson.M{"player_id": playerID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return default quests for new player
			return game.SimpleQuests(), nil
		}
		return nil, err
	}

	return result.Quests, nil
}

// GetAllOnlinePlayers returns all players currently online
func (db *MongoDB) GetAllOnlinePlayers(ctx context.Context) ([]*game.Entity, error) {
	collection := db.database.Collection("players")

	cursor, err := collection.Find(ctx, bson.M{"online": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []*game.Entity
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}

	return players, nil
}

// SetPlayerOnline sets a player's online status
func (db *MongoDB) SetPlayerOnline(ctx context.Context, playerID string, online bool) error {
	collection := db.database.Collection("players")

	filter := bson.M{"id": playerID}
	update := bson.M{
		"$set": bson.M{
			"online":     online,
			"last_seen":  time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}
