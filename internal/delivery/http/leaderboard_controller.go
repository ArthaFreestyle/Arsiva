package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type LeaderboardController interface {
	GetPublic(ctx fiber.Ctx) error
	GetGroup(ctx fiber.Ctx) error
}

type leaderboardControllerImpl struct {
	UseCase usecase.LeaderboardUseCase
	Log     *logrus.Logger
}

func NewLeaderboardController(uc usecase.LeaderboardUseCase, log *logrus.Logger) LeaderboardController {
	return &leaderboardControllerImpl{UseCase: uc, Log: log}
}

func (c *leaderboardControllerImpl) GetPublic(ctx fiber.Ctx) error {
	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "page tidak valid")
	}
	size, err := strconv.Atoi(ctx.Query("size", "20"))
	if err != nil || size < 1 || size > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "size tidak valid")
	}

	var sekolahId int
	if raw := ctx.Query("sekolah_id"); raw != "" {
		sekolahId, err = strconv.Atoi(raw)
		if err != nil || sekolahId < 1 {
			return fiber.NewError(fiber.StatusBadRequest, "sekolah_id tidak valid")
		}
	}

	req := &model.PublicLeaderboardRequest{
		Period:    ctx.Query("period", "alltime"),
		SekolahId: sekolahId,
		Page:      page,
		Size:      size,
	}

	resp, err := c.UseCase.GetPublic(ctx.Context(), req)
	if err != nil {
		c.Log.Warnf("GetPublic: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.PublicLeaderboardResponse]{Data: resp})
}

func (c *leaderboardControllerImpl) GetGroup(ctx fiber.Ctx) error {
	groupId := ctx.Params("id")

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "page tidak valid")
	}
	size, err := strconv.Atoi(ctx.Query("size", "20"))
	if err != nil || size < 1 || size > 100 {
		return fiber.NewError(fiber.StatusBadRequest, "size tidak valid")
	}

	claims := ctx.Locals("user").(*model.Claims)
	req := &model.GroupLeaderboardRequest{Page: page, Size: size}

	resp, err := c.UseCase.GetGroup(ctx.Context(), groupId, req, claims)
	if err != nil {
		c.Log.Warnf("GetGroup: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GroupLeaderboardResponse]{Data: resp})
}
