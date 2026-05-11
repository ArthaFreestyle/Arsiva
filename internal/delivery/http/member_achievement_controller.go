package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type MemberAchievementController interface {
	Create(ctx fiber.Ctx) error
	FindAllMine(ctx fiber.Ctx) error
	FindOne(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
}

type memberAchievementControllerImpl struct {
	AchievementUseCase usecase.MemberAchievementUseCase
	Log                *logrus.Logger
}

func NewMemberAchievementController(uc usecase.MemberAchievementUseCase, log *logrus.Logger) MemberAchievementController {
	return &memberAchievementControllerImpl{
		AchievementUseCase: uc,
		Log:                log,
	}
}

func (c *memberAchievementControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.MemberAchievementCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)

	result, err := c.AchievementUseCase.Create(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Failed create member achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.MemberAchievementResponse]{Data: result})
}

func (c *memberAchievementControllerImpl) FindAllMine(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)

	achievements, err := c.AchievementUseCase.FindAllMine(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("Failed get member achievements: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.MemberAchievementResponse]{Data: achievements})
}

func (c *memberAchievementControllerImpl) FindOne(ctx fiber.Ctx) error {
	achievementId := ctx.Params("achievement_id")
	claims := ctx.Locals("user").(*model.Claims)

	result, err := c.AchievementUseCase.FindOne(ctx.Context(), achievementId, claims)
	if err != nil {
		c.Log.Warnf("Failed get member achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberAchievementResponse]{Data: result})
}

func (c *memberAchievementControllerImpl) Delete(ctx fiber.Ctx) error {
	memberId := ctx.Params("member_id")
	achievementId := ctx.Params("achievement_id")
	claims := ctx.Locals("user").(*model.Claims)

	if err := c.AchievementUseCase.Delete(ctx.Context(), memberId, achievementId, claims); err != nil {
		c.Log.Warnf("Failed delete member achievement: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Member achievement deleted successfully"})
}
