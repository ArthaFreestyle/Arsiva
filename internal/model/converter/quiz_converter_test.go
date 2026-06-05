package converter

import (
	"encoding/json"
	"strings"
	"testing"

	"ArthaFreestyle/Arsiva/internal/entity"
)

// TestToPublicQuizResponseHidesAnswerKey guards the member-facing quiz view:
// it must never serialize per-option Score (the answer key) or per-question
// Poin (scoring metadata), while keeping the other fields a member needs.
func TestToPublicQuizResponseHidesAnswerKey(t *testing.T) {
	quiz := &entity.Quiz{
		QuizId: 12,
		Judul:  "Quiz Title",
		Soal: []*entity.Question{
			{
				PertanyaanId:   1,
				KuisId:         12,
				TeksPertanyaan: "What is 2+2?",
				Tipe:           "single_choice",
				Poin:           100,
				Urutan:         1,
				Pilihan: []*entity.Option{
					{JawabanId: 10, TeksJawaban: "3", Score: 0},
					{JawabanId: 11, TeksJawaban: "4", Score: 100},
				},
			},
		},
	}

	res := ToPublicQuizResponse(quiz)

	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal public quiz response: %v", err)
	}
	body := string(raw)

	if strings.Contains(body, "\"score\"") {
		t.Errorf("member quiz response must not contain score (answer key): %s", body)
	}
	if strings.Contains(body, "\"poin\"") {
		t.Errorf("member quiz response must not contain poin (scoring metadata): %s", body)
	}

	// The fields a member legitimately needs must still be present.
	for _, want := range []string{"\"jawaban_id\"", "\"teks_jawaban\"", "\"pertanyaan_id\"", "\"teks_pertanyaan\"", "\"tipe\"", "\"urutan\""} {
		if !strings.Contains(body, want) {
			t.Errorf("member quiz response missing required field %s: %s", want, body)
		}
	}
}
