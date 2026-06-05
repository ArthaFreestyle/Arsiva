package converter

import (
	"encoding/json"
	"strings"
	"testing"

	"ArthaFreestyle/Arsiva/internal/entity"
)

// TestToPublicCeritaResponseHidesScoringMetadata guards the member-facing story
// view: it must never serialize per-scene ending_point (answer key), ending_type
// (reveals the desirable ending), or urutan (authoring order), while keeping the
// fields the client needs to render and navigate scenes.
func TestToPublicCeritaResponseHidesScoringMetadata(t *testing.T) {
	cerita := &entity.CeritaInteraktif{
		CeritaId: 7,
		Judul:    "Story Title",
		Scenes: []*entity.Scene{
			{
				SceneId:   1,
				CeritaId:  7,
				SceneKey:  "start",
				SceneText: "You wake up in a dark room…",
				SceneChoices: []map[string]any{
					{"text": "Open the door", "next": "hallway"},
				},
				IsEnding:    false,
				EndingPoint: 0,
				EndingType:  "",
				Urutan:      1,
			},
			{
				SceneId:     2,
				CeritaId:    7,
				SceneKey:    "good_end",
				SceneText:   "You escaped!",
				IsEnding:    true,
				EndingPoint: 100,
				EndingType:  "good_ending",
				Urutan:      2,
			},
		},
	}

	res := ToPublicCeritaResponse(cerita)

	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal public cerita response: %v", err)
	}
	body := string(raw)

	for _, leak := range []string{"\"ending_point\"", "\"ending_type\"", "\"urutan\""} {
		if strings.Contains(body, leak) {
			t.Errorf("member story response must not contain %s: %s", leak, body)
		}
	}

	// The fields the client legitimately needs must still be present.
	for _, want := range []string{"\"scene_id\"", "\"cerita_id\"", "\"scene_key\"", "\"scene_text\"", "\"scene_choices\"", "\"is_ending\""} {
		if !strings.Contains(body, want) {
			t.Errorf("member story response missing required field %s: %s", want, body)
		}
	}

	// Scene order must be preserved (authored order from the repository).
	if len(res.Scenes) != 2 || res.Scenes[0].SceneId != 1 || res.Scenes[1].SceneId != 2 {
		t.Errorf("expected scenes preserved in authored order, got %+v", res.Scenes)
	}
}
