package model

// ==================== Requests ====================

type PublicLeaderboardRequest struct {
	Period    string `query:"period"`
	SekolahId int    `query:"sekolah_id"`
	Page      int    `query:"page"`
	Size      int    `query:"size"`
}

type GroupLeaderboardRequest struct {
	Page int `query:"page"`
	Size int `query:"size"`
}

// ==================== Responses ====================

type LeaderboardSekolah struct {
	SekolahId   int    `json:"sekolah_id"`
	NamaSekolah string `json:"nama_sekolah"`
}

type PublicLeaderboardItem struct {
	Rank       int                 `json:"rank"`
	MemberId   int                 `json:"member_id"`
	Username   string              `json:"username"`
	FotoProfil *string             `json:"foto_profil"`
	Level      int                 `json:"level"`
	TotalXP    int                 `json:"total_xp"`
	MonthlyXP  int                 `json:"monthly_xp"`
	Sekolah    *LeaderboardSekolah `json:"sekolah"`
}

type PublicLeaderboardResponse struct {
	Period      string                  `json:"period"`
	PeriodStart *string                 `json:"period_start"`
	Page        int                     `json:"page"`
	Size        int                     `json:"size"`
	Total       int                     `json:"total"`
	Items       []PublicLeaderboardItem `json:"items"`
}

type GroupLeaderboardHeader struct {
	GroupId        string  `json:"group_id"`
	GroupName      string  `json:"group_name"`
	GroupThumbnail *string `json:"group_thumbnail"`
}

type GroupLeaderboardItem struct {
	Rank           int     `json:"rank"`
	MemberId       int     `json:"member_id"`
	Username       string  `json:"username"`
	FotoProfil     *string `json:"foto_profil"`
	Level          int     `json:"level"`
	GroupXP        int     `json:"group_xp"`
	CompletedCount int     `json:"completed_count"`
	TotalXP        int     `json:"total_xp"`
}

type GroupLeaderboardResponse struct {
	Group *GroupLeaderboardHeader `json:"group"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
	Total int                     `json:"total"`
	Items []GroupLeaderboardItem  `json:"items"`
}
