package game

// QuestType represents the type of quest
type QuestType string

const (
	QuestTypeKill    QuestType = "kill"
	QuestTypeCollect QuestType = "collect"
	QuestTypeReach   QuestType = "reach"
)

// Quest represents a simple quest
type Quest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        QuestType `json:"type"`
	Target      string    `json:"target"`       // creature type or item name
	Required    int       `json:"required"`     // how many needed
	Current     int       `json:"current"`      // current progress
	Reward      QuestReward `json:"reward"`
	Completed   bool      `json:"completed"`
}

// QuestReward represents quest rewards
type QuestReward struct {
	Gold int    `json:"gold"`
	XP   int    `json:"xp"`
	Item *Item  `json:"item,omitempty"`
}

// SimpleQuests returns a set of simple starter quests
func SimpleQuests() []Quest {
	return []Quest{
		{
			ID:          "kill_goblins_5",
			Title:       "Goblin Slayer",
			Description: "Defeat 5 goblins",
			Type:        QuestTypeKill,
			Target:      "goblin",
			Required:    5,
			Current:     0,
			Reward: QuestReward{
				Gold: 50,
				XP:   100,
			},
			Completed: false,
		},
		{
			ID:          "reach_floor_5",
			Title:       "Deep Delver",
			Description: "Reach dungeon floor 5",
			Type:        QuestTypeReach,
			Target:      "floor",
			Required:    5,
			Current:     0,
			Reward: QuestReward{
				Gold: 100,
				XP:   200,
			},
			Completed: false,
		},
		{
			ID:          "collect_gold_100",
			Title:       "Treasure Hunter",
			Description: "Collect 100 gold",
			Type:        QuestTypeCollect,
			Target:      "gold",
			Required:    100,
			Current:     0,
			Reward: QuestReward{
				Gold: 50,
				XP:   150,
			},
			Completed: false,
		},
	}
}

// UpdateQuestProgress updates quest based on action
func UpdateQuestProgress(quests []Quest, questType QuestType, target string, amount int) []Quest {
	for i := range quests {
		if quests[i].Completed {
			continue
		}

		if quests[i].Type == questType && quests[i].Target == target {
			quests[i].Current += amount
			if quests[i].Current >= quests[i].Required {
				quests[i].Current = quests[i].Required
				quests[i].Completed = true
			}
		}
	}
	return quests
}

// GetActiveQuests returns all non-completed quests
func GetActiveQuests(quests []Quest) []Quest {
	var active []Quest
	for _, q := range quests {
		if !q.Completed {
			active = append(active, q)
		}
	}
	return active
}

// GetCompletedQuests returns all completed quests
func GetCompletedQuests(quests []Quest) []Quest {
	var completed []Quest
	for _, q := range quests {
		if q.Completed {
			completed = append(completed, q)
		}
	}
	return completed
}
