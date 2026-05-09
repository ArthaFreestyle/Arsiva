package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type GuruController interface {
	Create(ctx fiber.Ctx) error
	FindById(ctx fiber.Ctx) error
	FindAll(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
	GetMe(ctx fiber.Ctx) error
}

type guruControllerImpl struct {
	GuruUseCase usecase.GuruUseCase
	Log         *logrus.Logger
}

func NewGuruController(guruUseCase usecase.GuruUseCase, log *logrus.Logger) GuruController {
	return &guruControllerImpl{
		GuruUseCase: guruUseCase,
		Log:         log,
	}
}

func (c *guruControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.GuruCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	guru, err := c.GuruUseCase.Create(ctx.Context(), req)
	if err != nil {
		c.Log.Warnf("Failed create guru: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.GuruResponse]{Data: guru})
}

func (c *guruControllerImpl) FindById(ctx fiber.Ctx) error {
	guruId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	guru, err := c.GuruUseCase.FindById(ctx.Context(), guruId, claims)
	if err != nil {
		c.Log.Warnf("Failed get guru by id: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GuruDetailResponse]{Data: guru})
}

func (c *guruControllerImpl) FindAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	gurus, total, err := c.GuruUseCase.FindAll(ctx.Context(), search, page, size)
	if err != nil {
		c.Log.Warnf("Failed get all guru: %v", err)
		return err
	}

	_ = total
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.GuruResponse]{
		Data: gurus,
		Paging: &model.PageMetaData{
			Page: page,
			Size: size,
		},
	})
}

func (c *guruControllerImpl) Update(ctx fiber.Ctx) error {
	req := new(model.GuruUpdateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	guruId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)

	guru, err := c.GuruUseCase.Update(ctx.Context(), guruId, req, claims)
	if err != nil {
		c.Log.Warnf("Failed update guru: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GuruResponse]{Data: guru})
}

func (c *guruControllerImpl) Delete(ctx fiber.Ctx) error {
	guruId := ctx.Params("id")

	err := c.GuruUseCase.Delete(ctx.Context(), guruId)
	if err != nil {
		c.Log.Warnf("Failed delete guru: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Guru deleted successfully"})
}

func (c *guruControllerImpl) GetMe(ctx fiber.Ctx) error {
	claims := ctx.Locals("user").(*model.Claims)

	guru, err := c.GuruUseCase.GetMe(ctx.Context(), claims)
	if err != nil {
		c.Log.Warnf("Failed get guru profile: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.GuruDetailResponse]{Data: guru})
}
