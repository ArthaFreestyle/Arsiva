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
	GetAllPuzzle(ctx context.Context, page int, size int, search string) ([]*model.PuzzleResponse, int, error)
	GetPuzzleById(ctx context.Context, puzzleId string) (*model.PuzzleResponse, error)
	CreatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest,userId string) (*model.PuzzleResponse, error)
	UpdatePuzzle(ctx context.Context, puzzle *model.PuzzleRequest, puzzleId string) (*model.PuzzleResponse, error)
	DeletePuzzle(ctx context.Context, puzzleId string) (error)
}

type puzzleUseCaseImpl struct {
	PuzzleRepository repository.PuzzleRepository
	AssetRepository repository.AssetRepository
	Log *logrus.Logger
	Validator *validator.Validate
}

func NewPuzzleUseCase(puzzleRepository repository.PuzzleRepository, assetRepository repository.AssetRepository, log *logrus.Logger,validator *validator.Validate) PuzzleUseCase {
	return &puzzleUseCaseImpl{
		PuzzleRepository: puzzleRepository,
		AssetRepository: assetRepository,
		Log: log,
		Validator: validator,
	}
}

func (u *puzzleUseCaseImpl) GetAllPuzzle(ctx context.Context, page int, size int, search string) ([]*model.PuzzleResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	puzzles,total,err := u.PuzzleRepository.FindAll(ctx,page,size,search)
	if err != nil {
		u.Log.Warnf("error when get all puzzle: %v",err)
		return nil,0,fiber.ErrInternalServerError
	}

	res := converter.ToPuzzleResponses(puzzles)
	return res,total,nil
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
		GambarAssetId: puzzle.GambarAssetId,
		ThumbnailAssetId: puzzle.ThumbnailAssetId,
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

	var assetIds []int
	if puzzle.GambarAssetId != nil {
		assetIds = append(assetIds, *puzzle.GambarAssetId)
	}
	if puzzle.ThumbnailAssetId != nil {
		assetIds = append(assetIds, *puzzle.ThumbnailAssetId)
	}
	if len(assetIds) > 0 {
		if err := u.AssetRepository.MarkAsUsed(ctx, assetIds); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
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
		GambarAssetId: puzzle.GambarAssetId,
		ThumbnailAssetId: puzzle.ThumbnailAssetId,
		Kategori: puzzle.Kategori,
		XpReward: puzzle.XpReward,
		IsPublished: puzzle.IsPublished,
	}

	updatedPuzzle,err := u.PuzzleRepository.Update(ctx,UpdatedPuzzle)
	if err != nil {
		u.Log.Warnf("error when update puzzle: %v",err)
		return nil,fiber.ErrInternalServerError
	}

	var assetIds []int
	if puzzle.GambarAssetId != nil {
		assetIds = append(assetIds, *puzzle.GambarAssetId)
	}
	if puzzle.ThumbnailAssetId != nil {
		assetIds = append(assetIds, *puzzle.ThumbnailAssetId)
	}
	if len(assetIds) > 0 {
		if err := u.AssetRepository.MarkAsUsed(ctx, assetIds); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
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