package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type CeritaController interface {
	GetAllCerita(ctx fiber.Ctx) error
	GetCeritaById(ctx fiber.Ctx) error
	CreateCerita(ctx fiber.Ctx) error
	UpdateCerita(ctx fiber.Ctx) error
	CreateScene(ctx fiber.Ctx) error
	UpdateScene(ctx fiber.Ctx) error
	DeleteScene(ctx fiber.Ctx) error
	DeleteCerita(ctx fiber.Ctx) error

	GetAllCeritaManage(ctx fiber.Ctx) error
	GetCeritaByIdManage(ctx fiber.Ctx) error
}

type ceritaControllerImpl struct {
	CeritaUseCase usecase.CeritaUseCase
	Log           *logrus.Logger
}

func NewCeritaController(ceritaUseCase usecase.CeritaUseCase, log *logrus.Logger) CeritaController {
	return &ceritaControllerImpl{
		CeritaUseCase: ceritaUseCase,
		Log:           log,
	}
}

func (c *ceritaControllerImpl) GetAllCerita(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	ceritas, total, err := c.CeritaUseCase.GetAllCerita(ctx, page, size, search)
	if err != nil {
		c.Log.Warnf("error when get all cerita: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.CeritaResponse]{
		Data: ceritas,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) GetCeritaById(ctx fiber.Ctx) error {
	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}

	cerita, err := c.CeritaUseCase.GetCeritaById(ctx, ceritaId)
	if err != nil {
		c.Log.Warnf("error when get cerita by id: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}

	res := model.WebResponse[*model.CeritaResponse]{
		Data: cerita,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) CreateCerita(ctx fiber.Ctx) error {
	var cerita model.CeritaRequest
	if err := ctx.Bind().Body(&cerita); err != nil {
		c.Log.Warnf("error when bind cerita: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	userId := ctx.Locals("userId").(string)

	createdCerita, err := c.CeritaUseCase.CreateCerita(ctx, &cerita, userId)
	if err != nil {
		c.Log.Warnf("error when create cerita: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}

	res := model.WebResponse[*model.CeritaResponse]{
		Data: createdCerita,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) UpdateCerita(ctx fiber.Ctx) error {
	var cerita model.CeritaRequest
	if err := ctx.Bind().Body(&cerita); err != nil {
		c.Log.Warnf("error when bind cerita: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	updatedCerita, err := c.CeritaUseCase.UpdateCeritaManage(ctx, &cerita, ceritaId, userId, role)
	if err != nil {
		c.Log.Warnf("error when update cerita: %v", err)
		return err
	}

	res := model.WebResponse[*model.CeritaResponse]{
		Data: updatedCerita,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) DeleteCerita(ctx fiber.Ctx) error {
	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	err = c.CeritaUseCase.DeleteCeritaManage(ctx, ceritaId, userId, role)
	if err != nil {
		c.Log.Warnf("error when delete cerita: %v", err)
		return err
	}

	res := model.WebResponse[any]{
		Data: "cerita deleted successfully",
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) CreateScene(ctx fiber.Ctx) error {
	var scene model.SceneRequest
	if err := ctx.Bind().Body(&scene); err != nil {
		c.Log.Warnf("error when bind scene: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	createdScene, err := c.CeritaUseCase.CreateSceneManage(ctx, ceritaId, &scene, userId, role)
	if err != nil {
		c.Log.Warnf("error when create scene: %v", err)
		return err
	}

	res := model.WebResponse[*model.SceneResponse]{
		Data: createdScene,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) UpdateScene(ctx fiber.Ctx) error {
	var scene model.SceneRequest
	if err := ctx.Bind().Body(&scene); err != nil {
		c.Log.Warnf("error when bind scene: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}
	sceneId, err := strconv.Atoi(ctx.Params("scene_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	updatedScene, err := c.CeritaUseCase.UpdateSceneManage(ctx, ceritaId, sceneId, &scene, userId, role)
	if err != nil {
		c.Log.Warnf("error when update scene: %v", err)
		return err
	}

	res := model.WebResponse[*model.SceneResponse]{
		Data: updatedScene,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) DeleteScene(ctx fiber.Ctx) error {
	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}
	sceneId, err := strconv.Atoi(ctx.Params("scene_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	err = c.CeritaUseCase.DeleteSceneManage(ctx, ceritaId, sceneId, userId, role)
	if err != nil {
		c.Log.Warnf("error when delete scene: %v", err)
		return err
	}

	res := model.WebResponse[any]{
		Data: "scene deleted successfully",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) GetAllCeritaManage(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	ceritas, total, err := c.CeritaUseCase.GetAllCeritaManage(ctx, page, size, search, userId, role)
	if err != nil {
		c.Log.Warnf("error when get all cerita manage: %v", err)
		return err
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.CeritaResponse]{
		Data: ceritas,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *ceritaControllerImpl) GetCeritaByIdManage(ctx fiber.Ctx) error {
	ceritaId, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid cerita id")
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	cerita, err := c.CeritaUseCase.GetCeritaByIdManage(ctx, ceritaId, userId, role)
	if err != nil {
		c.Log.Warnf("error when get cerita by id manage: %v", err)
		return err
	}

	res := model.WebResponse[*model.CeritaResponse]{
		Data: cerita,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}
