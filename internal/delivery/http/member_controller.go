package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type MemberController interface {
	Create(ctx fiber.Ctx) error
	FindById(ctx fiber.Ctx) error
	FindAll(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
	GetMe(ctx fiber.Ctx) error
	UpdateMe(ctx fiber.Ctx) error
	GetProfile(ctx fiber.Ctx) error
}

type memberControllerImpl struct {
	MemberUseCase usecase.MemberUseCase
	Log           *logrus.Logger
}

func NewMemberController(memberUseCase usecase.MemberUseCase, log *logrus.Logger) MemberController {
	return &memberControllerImpl{
		MemberUseCase: memberUseCase,
		Log:           log,
	}
}

func (c *memberControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.MemberCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)

	member, err := c.MemberUseCase.Create(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Failed create member: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.MemberResponse]{Data: member})
}

func (c *memberControllerImpl) FindById(ctx fiber.Ctx) error {
	memberId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	member, err := c.MemberUseCase.FindById(ctx.Context(), memberId, claims)
	if err != nil {
		c.Log.Warnf("Failed get member by id: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberDetailResponse]{Data: member})
}

func (c *memberControllerImpl) FindAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	members, total, err := c.MemberUseCase.FindAll(ctx.Context(), search, page, size)
	if err != nil {
		c.Log.Warnf("Failed get all member: %v", err)
		return err
	}

	_ = total
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.MemberResponse]{
		Data: members,
		Paging: &model.PageMetaData{
			Page: page,
			Size: size,
		},
	})
}

func (c *memberControllerImpl) Update(ctx fiber.Ctx) error {
	req := new(model.MemberUpdateProfileRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	memberId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	member, err := c.MemberUseCase.Update(ctx.Context(), memberId, req, claims)
	if err != nil {
		c.Log.Warnf("Failed update member: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberResponse]{Data: member})
}

func (c *memberControllerImpl) Delete(ctx fiber.Ctx) error {
	memberId := ctx.Params("id")

	err := c.MemberUseCase.Delete(ctx.Context(), memberId)
	if err != nil {
		c.Log.Warnf("Failed delete member: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Member deleted successfully"})
}

func (c *memberControllerImpl) GetMe(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)

	member, err := c.MemberUseCase.GetMe(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("Failed get member profile: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberDetailResponse]{Data: member})
}

func (c *memberControllerImpl) UpdateMe(ctx fiber.Ctx) error {
	req := new(model.MemberUpdateProfileRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)

	member, err := c.MemberUseCase.UpdateMe(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Failed update member profile: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberResponse]{Data: member})
}

func (c *memberControllerImpl) GetProfile(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)

	profile, err := c.MemberUseCase.GetMyProfile(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("Failed get member full profile: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.MemberProfileResponse]{Data: profile})
}
