package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type AchievementController interface {
	Create(ctx fiber.Ctx) error
	FindById(ctx fiber.Ctx) error
	FindAll(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
}

type achievementControllerImpl struct {
	AchievementUseCase usecase.AchievementUseCase
	Log                *logrus.Logger
}

func NewAchievementController(achievementUseCase usecase.AchievementUseCase, log *logrus.Logger) AchievementController {
	return &achievementControllerImpl{
		AchievementUseCase: achievementUseCase,
		Log:                log,
	}
}

func (c *achievementControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.AchievementCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	achievement, err := c.AchievementUseCase.Create(ctx.Context(), req)
	if err != nil {
		c.Log.Warnf("Failed create achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.AchievementResponse]{Data: achievement})
}

func (c *achievementControllerImpl) FindById(ctx fiber.Ctx) error {
	achievementId := ctx.Params("id")

	achievement, err := c.AchievementUseCase.FindById(ctx.Context(), achievementId)
	if err != nil {
		c.Log.Warnf("Failed get achievement by id: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.AchievementResponse]{Data: achievement})
}

func (c *achievementControllerImpl) FindAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")
	tier := ctx.Query("tier", "")

	achievements, total, err := c.AchievementUseCase.FindAll(ctx.Context(), search, tier, page, size)
	if err != nil {
		c.Log.Warnf("Failed get all achievement: %v", err)
		return err
	}

	_ = total
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.AchievementResponse]{
		Data: achievements,
		Paging: &model.PageMetaData{
			Page: page,
			Size: size,
		},
	})
}

func (c *achievementControllerImpl) Update(ctx fiber.Ctx) error {
	req := new(model.AchievementUpdateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	achievementId := ctx.Params("id")

	achievement, err := c.AchievementUseCase.Update(ctx.Context(), achievementId, req)
	if err != nil {
		c.Log.Warnf("Failed update achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.AchievementResponse]{Data: achievement})
}

func (c *achievementControllerImpl) Delete(ctx fiber.Ctx) error {
	achievementId := ctx.Params("id")

	err := c.AchievementUseCase.Delete(ctx.Context(), achievementId)
	if err != nil {
		c.Log.Warnf("Failed delete achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Achievement deleted successfully"})
}
