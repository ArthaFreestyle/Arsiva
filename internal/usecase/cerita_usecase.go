package usecase

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/model/converter"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type CeritaUseCase interface {
	GetAllCerita(ctx context.Context, page int, size int, search string) ([]*model.CeritaResponse, int, error)
	GetCeritaById(ctx context.Context, ceritaId int) (*model.CeritaResponse, error)
	CreateCerita(ctx context.Context, cerita *model.CeritaRequest, userId string) (*model.CeritaResponse, error)
	UpdateCerita(ctx context.Context, cerita *model.CeritaRequest, ceritaId int) (*model.CeritaResponse, error)
	DeleteCerita(ctx context.Context, ceritaId int) error
}

type ceritaUseCaseImpl struct {
	CeritaRepository repository.CeritaRepository
	Log              *logrus.Logger
	Validator        *validator.Validate
}

func NewCeritaUseCase(ceritaRepository repository.CeritaRepository, log *logrus.Logger, validator *validator.Validate) CeritaUseCase {
	return &ceritaUseCaseImpl{
		CeritaRepository: ceritaRepository,
		Log:              log,
		Validator:        validator,
	}
}

func (u *ceritaUseCaseImpl) GetAllCerita(ctx context.Context, page int, size int, search string) ([]*model.CeritaResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	ceritas, total, err := u.CeritaRepository.FindAll(ctx, page, size, search)
	if err != nil {
		u.Log.Warnf("error when get all cerita: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	res := converter.ToCeritaResponses(ceritas)
	return res, total, nil
}

func (u *ceritaUseCaseImpl) GetCeritaById(ctx context.Context, ceritaId int) (*model.CeritaResponse, error) {
	cerita, err := u.CeritaRepository.FindById(ctx, ceritaId)
	if err != nil {
		u.Log.Warnf("error when get cerita by id: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	res := converter.ToCeritaResponse(cerita)
	return res, nil
}

func (u *ceritaUseCaseImpl) CreateCerita(ctx context.Context, cerita *model.CeritaRequest, userId string) (*model.CeritaResponse, error) {
	err := u.Validator.Struct(cerita)
	if err != nil {
		u.Log.Warnf("error when validate cerita: %v", err)
		return nil, fiber.ErrBadRequest
	}

	userIdInt, _ := strconv.Atoi(userId)

	// Build scenes
	scenes := make([]*entity.Scene, 0, len(cerita.Scenes))
	for i, s := range cerita.Scenes {
		choices := make([]map[string]interface{}, 0, len(s.SceneChoices))
		for _, c := range s.SceneChoices {
			choices = append(choices, map[string]interface{}{
				"text": c.Text,
				"next": c.Next,
			})
		}

		urutan := s.Urutan
		if urutan == 0 {
			urutan = i + 1
		}

		scenes = append(scenes, &entity.Scene{
			SceneKey:     s.SceneKey,
			SceneImage:   s.SceneImage,
			SceneText:    s.SceneText,
			SceneChoices: choices,
			IsEnding:     s.IsEnding,
			EndingPoint:  s.EndingPoint,
			EndingType:   s.EndingType,
			Urutan:       urutan,
		})
	}

	newCerita := &entity.CeritaInteraktif{
		Judul:     cerita.Judul,
		Thumbnail: cerita.Thumbnail,
		Deskripsi: cerita.Deskripsi,
		KategoriId: cerita.KategoriId,
		XpReward:  cerita.XpReward,
		CreatedBy: entity.User{
			UserId: strconv.Itoa(userIdInt),
		},
		IsPublished: cerita.IsPublished,
		Scenes:      scenes,
	}

	createdCerita, err := u.CeritaRepository.Create(ctx, newCerita)
	if err != nil {
		u.Log.Warnf("error when create cerita: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	res := converter.ToCeritaResponse(createdCerita)
	return res, nil
}

func (u *ceritaUseCaseImpl) UpdateCerita(ctx context.Context, cerita *model.CeritaRequest, ceritaId int) (*model.CeritaResponse, error) {
	err := u.Validator.Struct(cerita)
	if err != nil {
		u.Log.Warnf("error when validate cerita: %v", err)
		return nil, fiber.ErrBadRequest
	}

	// Build scenes
	scenes := make([]*entity.Scene, 0, len(cerita.Scenes))
	for i, s := range cerita.Scenes {
		choices := make([]map[string]interface{}, 0, len(s.SceneChoices))
		for _, c := range s.SceneChoices {
			choices = append(choices, map[string]interface{}{
				"text": c.Text,
				"next": c.Next,
			})
		}

		urutan := s.Urutan
		if urutan == 0 {
			urutan = i + 1
		}

		scenes = append(scenes, &entity.Scene{
			SceneKey:     s.SceneKey,
			SceneImage:   s.SceneImage,
			SceneText:    s.SceneText,
			SceneChoices: choices,
			IsEnding:     s.IsEnding,
			EndingPoint:  s.EndingPoint,
			EndingType:   s.EndingType,
			Urutan:       urutan,
		})
	}

	updatedCerita := &entity.CeritaInteraktif{
		CeritaId:    ceritaId,
		Judul:       cerita.Judul,
		Thumbnail:   cerita.Thumbnail,
		Deskripsi:   cerita.Deskripsi,
		KategoriId:  cerita.KategoriId,
		XpReward:    cerita.XpReward,
		IsPublished: cerita.IsPublished,
		Scenes:      scenes,
	}

	result, err := u.CeritaRepository.Update(ctx, updatedCerita)
	if err != nil {
		u.Log.Warnf("error when update cerita: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	res := converter.ToCeritaResponse(result)
	return res, nil
}

func (u *ceritaUseCaseImpl) DeleteCerita(ctx context.Context, ceritaId int) error {
	err := u.CeritaRepository.Delete(ctx, ceritaId)
	if err != nil {
		u.Log.Warnf("error when delete cerita: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}
