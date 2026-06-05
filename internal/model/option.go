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

// PublicOptionResponse is the member-facing option shape. It deliberately
// omits Score, which encodes the answer key (the highest-scoring option is
// the correct one), so members never receive it.
type PublicOptionResponse struct {
	JawabanId   int    `json:"jawaban_id"`
	TeksJawaban string `json:"teks_jawaban"`
}
