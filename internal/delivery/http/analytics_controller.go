package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type AnalyticsController interface {
	GetUserSummary(ctx fiber.Ctx) error
	GetContentSummary(ctx fiber.Ctx) error
	GetUserGrowth(ctx fiber.Ctx) error
	GetContentGrowth(ctx fiber.Ctx) error
}

type analyticsControllerImpl struct {
	UseCase usecase.AnalyticsUseCase
	Log     *logrus.Logger
}

func NewAnalyticsController(uc usecase.AnalyticsUseCase, log *logrus.Logger) AnalyticsController {
	return &analyticsControllerImpl{UseCase: uc, Log: log}
}

func (c *analyticsControllerImpl) GetUserSummary(ctx fiber.Ctx) error {
	resp, err := c.UseCase.GetUserSummary(ctx.Context())
	if err != nil {
		c.Log.Warnf("GetUserSummary: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.UserSummaryResponse]{Data: resp})
}

func (c *analyticsControllerImpl) GetContentSummary(ctx fiber.Ctx) error {
	resp, err := c.UseCase.GetContentSummary(ctx.Context())
	if err != nil {
		c.Log.Warnf("GetContentSummary: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ContentSummaryResponse]{Data: resp})
}

func (c *analyticsControllerImpl) GetUserGrowth(ctx fiber.Ctx) error {
	resp, err := c.UseCase.GetUserGrowth(ctx.Context(), parseGrowthQuery(ctx))
	if err != nil {
		c.Log.Warnf("GetUserGrowth: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.UserGrowthResponse]{Data: resp})
}

func (c *analyticsControllerImpl) GetContentGrowth(ctx fiber.Ctx) error {
	resp, err := c.UseCase.GetContentGrowth(ctx.Context(), parseGrowthQuery(ctx))
	if err != nil {
		c.Log.Warnf("GetContentGrowth: %v", err)
		return err
	}
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ContentGrowthResponse]{Data: resp})
}

// parseGrowthQuery only collects the raw params; validation, defaults and
// bucket snapping live in the usecase.
func parseGrowthQuery(ctx fiber.Ctx) *model.GrowthRequest {
	return &model.GrowthRequest{
		Interval: ctx.Query("interval"),
		From:     ctx.Query("from"),
		To:       ctx.Query("to"),
	}
}
