package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type PuzzleUseCase interface {
	GetAllPuzzle(ctx context.Context) ([]*model.PuzzleResponse, error)
	GetPuzzleById(ctx context.Context, puzzleId string) (*model.PuzzleResponse, error)
	CreatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest,userId string) (*model.PuzzleResponse, error)
	UpdatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest, puzzleId string) (*model.PuzzleResponse, error)
	DeletePuzzle(ctx context.Context, puzzleId string) (error)
}

type puzzleUseCaseImpl struct {
	PuzzleRepository repository.PuzzleRepository
	Log *logrus.Logger
	Validator *validator.Validate
}

func NewPuzzleUseCase(puzzleRepository repository.PuzzleRepository,log *logrus.Logger,validator *validator.Validate) PuzzleUseCase {
	return &puzzleUseCaseImpl{
		PuzzleRepository: puzzleRepository,
		Log: log,
		Validator: validator,
	}
}

func (u *puzzleUseCaseImpl) GetAllPuzzle(ctx context.Context) ([]*model.PuzzleResponse, error) {
	puzzles,err := u.PuzzleRepository.FindAll(ctx)
	if err != nil {
		u.Log.Warnf("error when get all puzzle: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToPuzzleResponses(puzzles)
	return res,nil
}

func (u *puzzleUseCaseImpl) GetPuzzleById(ctx context.Context, puzzleId string) (*model.PuzzleResponse, error) {
	puzzle,err := u.PuzzleRepository.FindById(ctx,puzzleId)
	if err != nil {
		u.Log.Warnf("error when get puzzle by id: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToPuzzleResponse(puzzle)
	return res,nil
}

func (u *puzzleUseCaseImpl) CreatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest,userId string) (*model.PuzzleResponse, error) {
	err := u.Validator.Struct(puzzle)
	if err != nil {
		u.Log.Warnf("error when validate puzzle: %v",err)
		return nil,fiber.ErrBadRequest
	}

	NewPuzzle := &entity.Puzzle{
		Judul: puzzle.Judul,
		Gambar: puzzle.Gambar,
		Thumbnail: puzzle.Thumbnail,
		Kategori: puzzle.Kategori,
		XpReward: puzzle.XpReward,
		CreatedBy: entity.User{
			UserId: userId,
		},
		IsPublished: puzzle.IsPublished,
	}

	createdPuzzle,err := u.PuzzleRepository.Create(ctx,NewPuzzle)
	if err != nil {
		u.Log.Warnf("error when create puzzle: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToPuzzleResponse(createdPuzzle)
	return res,nil
}

func (u *puzzleUseCaseImpl) UpdatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest, puzzleId string) (*model.PuzzleResponse, error) {
	err := u.Validator.Struct(puzzle)
	if err != nil {
		u.Log.Warnf("error when validate puzzle: %v",err)
		return nil,fiber.ErrBadRequest
	}

	UpdatedPuzzle := &entity.Puzzle{
		PuzzleId: puzzleId,
		Judul: puzzle.Judul,
		Gambar: puzzle.Gambar,
		Thumbnail: puzzle.Thumbnail,
		Kategori: puzzle.Kategori,
		XpReward: puzzle.XpReward,
		IsPublished: puzzle.IsPublished,
	}

	updatedPuzzle,err := u.PuzzleRepository.Update(ctx,UpdatedPuzzle)
	if err != nil {
		u.Log.Warnf("error when update puzzle: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	res := converter.ToPuzzleResponse(updatedPuzzle)
	return res,nil
}

func (u *puzzleUseCaseImpl) DeletePuzzle(ctx context.Context, puzzleId string) (error) {
	err := u.PuzzleRepository.Delete(ctx,puzzleId)
	if err != nil {
		u.Log.Warnf("error when delete puzzle: %v",err)
		return fiber.ErrInternalServerError
	}
	return nil
}