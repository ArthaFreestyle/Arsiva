package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type GamificationController interface {
	GetMyStreak(ctx fiber.Ctx) error
	GetTodayTasks(ctx fiber.Ctx) error
}

type gamificationControllerImpl struct {
	UseCase usecase.GamificationUseCase
	Log     *logrus.Logger
}

func NewGamificationController(uc usecase.GamificationUseCase, log *logrus.Logger) GamificationController {
	return &gamificationControllerImpl{UseCase: uc, Log: log}
}

func (c *gamificationControllerImpl) GetMyStreak(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.GetStreak(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("GetMyStreak: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.StreakResponse]{Data: resp})
}

func (c *gamificationControllerImpl) GetTodayTasks(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.GetTodayTasks(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("GetTodayTasks: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.DailyTasksResponse]{Data: resp})
}
