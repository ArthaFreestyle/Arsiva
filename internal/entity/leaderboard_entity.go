package entity

type LeaderboardEntry struct {
	Rank       int     `db:"rank"`
	MemberId   int     `db:"member_id"`
	Username   string  `db:"username"`
	FotoProfil *string `db:"foto_profil"`
	Level      int     `db:"level"`
	TotalXP    int     `db:"total_xp"`
	// Public-monthly only. Zero for alltime / group.
	MonthlyXP int `db:"monthly_xp"`
	// Group-only fields. Zero for public leaderboard.
	GroupXP        int `db:"group_xp"`
	CompletedCount int `db:"completed_count"`
	// Public-only — nil for group leaderboard.
	SekolahId   *int    `db:"sekolah_id"`
	SekolahNama *string `db:"nama_sekolah"`
	// Window function total — same value on every row, read once.
	TotalCount int `db:"total_count"`
}
