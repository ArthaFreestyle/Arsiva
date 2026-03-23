package converter

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
)

func ToPuzzleResponse(puzzle *entity.Puzzle) *model.PuzzleResponse {
	return &model.PuzzleResponse{
		PuzzleId: puzzle.PuzzleId,
		Judul: puzzle.Judul,
		Gambar: puzzle.Gambar,
		Kategori: puzzle.Kategori,
		XpReward: puzzle.XpReward,
		CreatedBy: *ToUserResponse(&puzzle.CreatedBy),
		CreatedAt: puzzle.CreatedAt,
		IsPublished: puzzle.IsPublished,
	}
}

func ToPuzzleResponses(puzzles []*entity.Puzzle) []*model.PuzzleResponse {
	var responses []*model.PuzzleResponse
	for _, puzzle := range puzzles {
		responses = append(responses, ToPuzzleResponse(puzzle))
	}
	return responses
}