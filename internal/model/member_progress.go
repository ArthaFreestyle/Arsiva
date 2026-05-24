package model

// ==================== Requests ====================

type ProgressStartRequest struct {
	ContentType     string `json:"content_type"      validate:"required,oneof=kuis cerita puzzle"`
	ContentId       int    `json:"content_id"        validate:"required,gt=0"`
	GroupId         string `json:"group_id"`
	DurationSeconds int    `json:"duration_seconds"  validate:"required,min=10,max=3600"`
}

type ProgressAnswerRequest struct {
	ContentType  string `json:"content_type"   validate:"required,oneof=kuis"`
	ContentId    int    `json:"content_id"     validate:"required,gt=0"`
	PertanyaanId int    `json:"pertanyaan_id"  validate:"required,gt=0"`
	JawabanId    int    `json:"jawaban_id"     validate:"required,gt=0"`
}

type ProgressSceneRequest struct {
	ContentType string `json:"content_type"  validate:"required,oneof=cerita"`
	ContentId   int    `json:"content_id"    validate:"required,gt=0"`
	SceneId     int    `json:"scene_id"      validate:"required,gt=0"`
}

type ProgressSolveRequest struct {
	ContentType string `json:"content_type"  validate:"required,oneof=puzzle"`
	ContentId   int    `json:"content_id"    validate:"required,gt=0"`
	Solved      bool   `json:"solved"`
}

type ProgressSubmitRequest struct {
	ContentType string `json:"content_type"  validate:"required,oneof=kuis cerita puzzle"`
	ContentId   int    `json:"content_id"    validate:"required,gt=0"`
}

// ==================== Responses ====================

type ProgressStartResponse struct {
	SessionKey string `json:"session_key"`
	ExpiresAt  int64  `json:"expires_at"`
	MaxScore   int    `json:"max_score"`
}

type ProgressAnswerResponse struct {
	RunningSkor int `json:"running_skor"`
}

type ProgressSessionResponse struct {
	SessionKey      string         `json:"session_key"`
	MemberId        string         `json:"member_id"`
	GroupId         string         `json:"group_id"`
	ContentType     string         `json:"content_type"`
	ContentId       int            `json:"content_id"`
	ExpiresAt       int64          `json:"expires_at"`
	DurationSeconds int            `json:"duration_seconds"`
	MaxScore        int            `json:"max_score"`
	RunningSkor     int            `json:"running_skor"`
	State           string         `json:"state"`
	Answers         map[string]int `json:"answers"`
}

type ProgressFinalizeResponse struct {
	ProgresId     int  `json:"progres_id"`
	LeveledUp     bool `json:"leveled_up"`
	PreviousLevel int  `json:"previous_level"`
	NewLevel      int  `json:"new_level"`
}
