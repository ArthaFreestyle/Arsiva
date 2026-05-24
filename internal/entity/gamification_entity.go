package entity

import "time"

type MemberStreak struct {
	MemberId         int        `db:"member_id"`
	CurrentStreak    int        `db:"current_streak"`
	LongestStreak    int        `db:"longest_streak"`
	LastActiveDate   *time.Time `db:"last_active_date"`
	FreezesAvailable int        `db:"freezes_available"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

type DailyTask struct {
	DailyTaskId  int        `db:"daily_task_id"`
	MemberId     int        `db:"member_id"`
	TaskDate     time.Time  `db:"task_date"`
	TaskType     string     `db:"task_type"`
	TargetCount  int        `db:"target_count"`
	CurrentCount int        `db:"current_count"`
	XpReward     int        `db:"xp_reward"`
	CompletedAt  *time.Time `db:"completed_at"`
	CreatedAt    time.Time  `db:"created_at"`
}
