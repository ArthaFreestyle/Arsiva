package model

// ─── Summary ──────────────────────────────────────────────────────────────────

type UserRoleCountResponse struct {
	SuperAdmin int `json:"super_admin"`
	Guru       int `json:"guru"`
	Member     int `json:"member"`
}

type UserSummaryResponse struct {
	Total  int                   `json:"total"`
	ByRole UserRoleCountResponse `json:"by_role"`
}

type PublishCountResponse struct {
	Total     int `json:"total"`
	Published int `json:"published"`
	Draft     int `json:"draft"`
}

type GameSummaryResponse struct {
	Total     int                  `json:"total"`
	Published int                  `json:"published"`
	Draft     int                  `json:"draft"`
	Kuis      PublishCountResponse `json:"kuis"`
	Cerita    PublishCountResponse `json:"cerita"`
	Puzzle    PublishCountResponse `json:"puzzle"`
}

type ContentSummaryResponse struct {
	Game    GameSummaryResponse  `json:"game"`
	Artikel PublishCountResponse `json:"artikel"`
}

// ─── Growth ───────────────────────────────────────────────────────────────────

// GrowthRequest carries the raw query params; the usecase validates them.
type GrowthRequest struct {
	Interval string
	From     string
	To       string
}

type UserGrowthPoint struct {
	PeriodStart string                `json:"period_start"`
	NewUsers    int                   `json:"new_users"`
	NewByRole   UserRoleCountResponse `json:"new_by_role"`
	Cumulative  int                   `json:"cumulative"`
}

type UserGrowthResponse struct {
	Interval string            `json:"interval"`
	From     string            `json:"from"`
	To       string            `json:"to"`
	Series   []UserGrowthPoint `json:"series"`
}

type ContentTypeCountResponse struct {
	Kuis    int `json:"kuis"`
	Cerita  int `json:"cerita"`
	Puzzle  int `json:"puzzle"`
	Artikel int `json:"artikel"`
	Total   int `json:"total"`
}

type ContentGrowthPoint struct {
	PeriodStart string                   `json:"period_start"`
	New         ContentTypeCountResponse `json:"new"`
	Cumulative  ContentTypeCountResponse `json:"cumulative"`
}

type ContentGrowthResponse struct {
	Interval string               `json:"interval"`
	From     string               `json:"from"`
	To       string               `json:"to"`
	Series   []ContentGrowthPoint `json:"series"`
}
