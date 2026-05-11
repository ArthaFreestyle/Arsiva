package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type MemberSocialLinkController interface {
	Create(ctx fiber.Ctx) error
	FindAllMine(ctx fiber.Ctx) error
	FindById(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
}

type memberSocialLinkControllerImpl struct {
	SocialLinkUseCase usecase.MemberSocialLinkUseCase
	Log               *logrus.Logger
}

func NewMemberSocialLinkController(socialLinkUseCase usecase.MemberSocialLinkUseCase, log *logrus.Logger) MemberSocialLinkController {
	return &memberSocialLinkControllerImpl{
		SocialLinkUseCase: socialLinkUseCase,
		Log:               log,
	}
}

func (c *memberSocialLinkControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.MemberSocialLinkCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)

	link, err := c.SocialLinkUseCase.Create(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Failed create social link: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.MemberSocialLinkResponse]{Data: link})
}

func (c *memberSocialLinkControllerImpl) FindAllMine(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)

	links, err := c.SocialLinkUseCase.FindAllMine(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("Failed get social links: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.MemberSocialLinkResponse]{Data: links})
}

func (c *memberSocialLinkControllerImpl) FindById(ctx fiber.Ctx) error {
	socialId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	link, err := c.SocialLinkUseCase.FindById(ctx.Context(), socialId, claims)
	if err != nil {
		c.Log.Warnf("Failed get social link by id: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberSocialLinkResponse]{Data: link})
}

func (c *memberSocialLinkControllerImpl) Update(ctx fiber.Ctx) error {
	req := new(model.MemberSocialLinkUpdateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	socialId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	link, err := c.SocialLinkUseCase.Update(ctx.Context(), socialId, req, claims)
	if err != nil {
		c.Log.Warnf("Failed update social link: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberSocialLinkResponse]{Data: link})
}

func (c *memberSocialLinkControllerImpl) Delete(ctx fiber.Ctx) error {
	socialId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	if err := c.SocialLinkUseCase.Delete(ctx.Context(), socialId, claims); err != nil {
		c.Log.Warnf("Failed delete social link: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Social link deleted successfully"})
}
