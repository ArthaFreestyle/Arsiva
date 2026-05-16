package entity

import "time"

type MemberProgress struct {
	ProgresId   int        `db:"progres_id"`
	MemberId    int        `db:"member_id"`
	GroupId     string     `db:"group_id"`
	ContentType string     `db:"content_type"`
	ContentId   int        `db:"content_id"`
	Skor        int        `db:"skor"`
	XpReward    int        `db:"xp_reward"`
	CompletedAt *time.Time `db:"completed_at"`
	Duration    int        `db:"duration"`
}
