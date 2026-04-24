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
	CreateScene(ctx context.Context, ceritaId int, scene *model.SceneRequest) (*model.SceneResponse, error)
	UpdateScene(ctx context.Context, ceritaId int, sceneId int, scene *model.SceneRequest) (*model.SceneResponse, error)
	DeleteScene(ctx context.Context, ceritaId int, sceneId int) error
	DeleteCerita(ctx context.Context, ceritaId int) error

	GetAllCeritaManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*model.CeritaResponse, int, error)
	GetCeritaByIdManage(ctx context.Context, ceritaId int, userId string, role string) (*model.CeritaResponse, error)
	UpdateCeritaManage(ctx context.Context, cerita *model.CeritaRequest, ceritaId int, userId string, role string) (*model.CeritaResponse, error)
	DeleteCeritaManage(ctx context.Context, ceritaId int, userId string, role string) error
	CreateSceneManage(ctx context.Context, ceritaId int, scene *model.SceneRequest, userId string, role string) (*model.SceneResponse, error)
	UpdateSceneManage(ctx context.Context, ceritaId int, sceneId int, scene *model.SceneRequest, userId string, role string) (*model.SceneResponse, error)
	DeleteSceneManage(ctx context.Context, ceritaId int, sceneId int, userId string, role string) error
}

type ceritaUseCaseImpl struct {
	CeritaRepository repository.CeritaRepository
	AssetRepository  repository.AssetRepository
	Log              *logrus.Logger
	Validator        *validator.Validate
}

func NewCeritaUseCase(ceritaRepository repository.CeritaRepository, assetRepository repository.AssetRepository, log *logrus.Logger, validator *validator.Validate) CeritaUseCase {
	return &ceritaUseCaseImpl{
		CeritaRepository: ceritaRepository,
		AssetRepository:  assetRepository,
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

	newCerita := &entity.CeritaInteraktif{
		Judul:            cerita.Judul,
		ThumbnailAssetId: cerita.ThumbnailAssetId,
		Deskripsi:        cerita.Deskripsi,
		KategoriId:       cerita.KategoriId,
		XpReward:         cerita.XpReward,
		CreatedBy: entity.User{
			UserId: strconv.Itoa(userIdInt),
		},
		IsPublished: false,
	}

	createdCerita, err := u.CeritaRepository.Create(ctx, newCerita)
	if err != nil {
		u.Log.Warnf("error when create cerita: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	var assetIds []int
	if cerita.ThumbnailAssetId != nil {
		assetIds = append(assetIds, *cerita.ThumbnailAssetId)
	}
	if len(assetIds) > 0 {
		if err := u.AssetRepository.MarkAsUsed(ctx, assetIds); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
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

	updatedCerita := &entity.CeritaInteraktif{
		CeritaId:         ceritaId,
		Judul:            cerita.Judul,
		ThumbnailAssetId: cerita.ThumbnailAssetId,
		Deskripsi:        cerita.Deskripsi,
		KategoriId:       cerita.KategoriId,
		XpReward:         cerita.XpReward,
		IsPublished:      cerita.IsPublished,
	}

	result, err := u.CeritaRepository.Update(ctx, updatedCerita)
	if err != nil {
		u.Log.Warnf("error when update cerita: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	var assetIds []int
	if cerita.ThumbnailAssetId != nil {
		assetIds = append(assetIds, *cerita.ThumbnailAssetId)
	}
	if len(assetIds) > 0 {
		if err := u.AssetRepository.MarkAsUsed(ctx, assetIds); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	res := converter.ToCeritaResponse(result)
	return res, nil
}

func (u *ceritaUseCaseImpl) CreateScene(ctx context.Context, ceritaId int, scene *model.SceneRequest) (*model.SceneResponse, error) {
	err := u.Validator.Struct(scene)
	if err != nil {
		u.Log.Warnf("error when validate scene: %v", err)
		return nil, fiber.ErrBadRequest
	}

	result, err := u.CeritaRepository.CreateScene(ctx, ceritaId, toSceneEntity(scene))
	if err != nil {
		u.Log.Warnf("error when create scene: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	if scene.SceneImageAssetId != nil {
		if err := u.AssetRepository.MarkAsUsed(ctx, []int{*scene.SceneImageAssetId}); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	return converter.ToSceneResponse(result), nil
}

func (u *ceritaUseCaseImpl) UpdateScene(ctx context.Context, ceritaId int, sceneId int, scene *model.SceneRequest) (*model.SceneResponse, error) {
	err := u.Validator.Struct(scene)
	if err != nil {
		u.Log.Warnf("error when validate scene: %v", err)
		return nil, fiber.ErrBadRequest
	}

	result, err := u.CeritaRepository.UpdateScene(ctx, ceritaId, sceneId, toSceneEntity(scene))
	if err != nil {
		u.Log.Warnf("error when update scene: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	if scene.SceneImageAssetId != nil {
		if err := u.AssetRepository.MarkAsUsed(ctx, []int{*scene.SceneImageAssetId}); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	return converter.ToSceneResponse(result), nil
}

func (u *ceritaUseCaseImpl) DeleteScene(ctx context.Context, ceritaId int, sceneId int) error {
	err := u.CeritaRepository.DeleteScene(ctx, ceritaId, sceneId)
	if err != nil {
		u.Log.Warnf("error when delete scene: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (u *ceritaUseCaseImpl) DeleteCerita(ctx context.Context, ceritaId int) error {
	err := u.CeritaRepository.Delete(ctx, ceritaId)
	if err != nil {
		u.Log.Warnf("error when delete cerita: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (u *ceritaUseCaseImpl) GetAllCeritaManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*model.CeritaResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	ceritas, total, err := u.CeritaRepository.FindAllManage(ctx, page, size, search, userId, role)
	if err != nil {
		u.Log.Warnf("error when get all cerita manage: %v", err)
		return nil, 0, fiber.ErrInternalServerError
	}

	res := converter.ToCeritaResponses(ceritas)
	return res, total, nil
}

func (u *ceritaUseCaseImpl) GetCeritaByIdManage(ctx context.Context, ceritaId int, userId string, role string) (*model.CeritaResponse, error) {
	cerita, err := u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("error when get cerita by id manage: %v", err)
		return nil, fiber.ErrNotFound
	}

	res := converter.ToCeritaResponse(cerita)
	return res, nil
}

func (u *ceritaUseCaseImpl) UpdateCeritaManage(ctx context.Context, cerita *model.CeritaRequest, ceritaId int, userId string, role string) (*model.CeritaResponse, error) {
	err := u.Validator.Struct(cerita)
	if err != nil {
		u.Log.Warnf("error when validate cerita: %v", err)
		return nil, fiber.ErrBadRequest
	}

	_, err = u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("cerita not found or access denied: %v", err)
		return nil, fiber.ErrForbidden
	}

	updatedCerita := &entity.CeritaInteraktif{
		CeritaId:         ceritaId,
		Judul:            cerita.Judul,
		ThumbnailAssetId: cerita.ThumbnailAssetId,
		Deskripsi:        cerita.Deskripsi,
		KategoriId:       cerita.KategoriId,
		XpReward:         cerita.XpReward,
		IsPublished:      cerita.IsPublished,
	}

	result, err := u.CeritaRepository.Update(ctx, updatedCerita)
	if err != nil {
		u.Log.Warnf("error when update cerita: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	var assetIds []int
	if cerita.ThumbnailAssetId != nil {
		assetIds = append(assetIds, *cerita.ThumbnailAssetId)
	}
	if len(assetIds) > 0 {
		if err := u.AssetRepository.MarkAsUsed(ctx, assetIds); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	res := converter.ToCeritaResponse(result)
	return res, nil
}

func (u *ceritaUseCaseImpl) DeleteCeritaManage(ctx context.Context, ceritaId int, userId string, role string) error {
	_, err := u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("cerita not found or access denied: %v", err)
		return fiber.ErrForbidden
	}

	err = u.CeritaRepository.Delete(ctx, ceritaId)
	if err != nil {
		u.Log.Warnf("error when delete cerita: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func (u *ceritaUseCaseImpl) CreateSceneManage(ctx context.Context, ceritaId int, scene *model.SceneRequest, userId string, role string) (*model.SceneResponse, error) {
	err := u.Validator.Struct(scene)
	if err != nil {
		u.Log.Warnf("error when validate scene: %v", err)
		return nil, fiber.ErrBadRequest
	}

	_, err = u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("cerita not found or access denied for scene create: %v", err)
		return nil, fiber.ErrForbidden
	}

	result, err := u.CeritaRepository.CreateScene(ctx, ceritaId, toSceneEntity(scene))
	if err != nil {
		u.Log.Warnf("error when create scene: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	if scene.SceneImageAssetId != nil {
		if err := u.AssetRepository.MarkAsUsed(ctx, []int{*scene.SceneImageAssetId}); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	return converter.ToSceneResponse(result), nil
}

func (u *ceritaUseCaseImpl) UpdateSceneManage(ctx context.Context, ceritaId int, sceneId int, scene *model.SceneRequest, userId string, role string) (*model.SceneResponse, error) {
	err := u.Validator.Struct(scene)
	if err != nil {
		u.Log.Warnf("error when validate scene: %v", err)
		return nil, fiber.ErrBadRequest
	}

	_, err = u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("cerita not found or access denied for scene update: %v", err)
		return nil, fiber.ErrForbidden
	}

	result, err := u.CeritaRepository.UpdateScene(ctx, ceritaId, sceneId, toSceneEntity(scene))
	if err != nil {
		u.Log.Warnf("error when update scene: %v", err)
		return nil, fiber.ErrInternalServerError
	}

	if scene.SceneImageAssetId != nil {
		if err := u.AssetRepository.MarkAsUsed(ctx, []int{*scene.SceneImageAssetId}); err != nil {
			u.Log.Warnf("failed to mark asset as used: %v", err)
		}
	}

	return converter.ToSceneResponse(result), nil
}

func (u *ceritaUseCaseImpl) DeleteSceneManage(ctx context.Context, ceritaId int, sceneId int, userId string, role string) error {
	_, err := u.CeritaRepository.FindByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		u.Log.Warnf("cerita not found or access denied for scene delete: %v", err)
		return fiber.ErrForbidden
	}

	err = u.CeritaRepository.DeleteScene(ctx, ceritaId, sceneId)
	if err != nil {
		u.Log.Warnf("error when delete scene: %v", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func toSceneEntity(scene *model.SceneRequest) *entity.Scene {
	choices := make([]map[string]interface{}, 0, len(scene.SceneChoices))
	for _, c := range scene.SceneChoices {
		choices = append(choices, map[string]interface{}{
			"text": c.Text,
			"next": c.Next,
		})
	}

	return &entity.Scene{
		SceneKey:          scene.SceneKey,
		SceneImageAssetId: scene.SceneImageAssetId,
		SceneText:         scene.SceneText,
		SceneChoices:      choices,
		IsEnding:          scene.IsEnding,
		EndingPoint:       scene.EndingPoint,
		EndingType:        scene.EndingType,
		Urutan:            scene.Urutan,
	}
}
