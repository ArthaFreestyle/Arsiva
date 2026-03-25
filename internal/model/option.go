package model

type OptionRequest struct {
	TeksJawaban string `json:"teks_jawaban" validate:"required"`
	Score       int    `json:"score"`
}

type OptionResponse struct {
	JawabanId   int    `json:"jawaban_id"`
	TeksJawaban string `json:"teks_jawaban"`
	Score       int    `json:"score"`
}
