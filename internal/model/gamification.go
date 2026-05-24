package model

type StreakResponse struct {
	CurrentStreak    int    `json:"current_streak"`
	LongestStreak    int    `json:"longest_streak"`
	FreezesAvailable int    `json:"freezes_available"`
	LastActiveDate   string `json:"last_active_date"` // "2006-01-02" or ""
	ActiveToday      bool   `json:"active_today"`
}

type DailyTaskItem struct {
	TaskType     string `json:"task_type"`
	TargetCount  int    `json:"target_count"`
	CurrentCount int    `json:"current_count"`
	XpReward     int    `json:"xp_reward"`
	Completed    bool   `json:"completed"`
}

type DailyTasksResponse struct {
	TaskDate     string          `json:"task_date"`
	Tasks        []DailyTaskItem `json:"tasks"`
	AllCompleted bool            `json:"all_completed"`
}
