package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type SekolahController interface {
	Create(ctx fiber.Ctx) error
	FindById(ctx fiber.Ctx) error
	FindAll(ctx fiber.Ctx) error
	Update(ctx fiber.Ctx) error
	Delete(ctx fiber.Ctx) error
}

type sekolahControllerImpl struct {
	SekolahUseCase usecase.SekolahUseCase
	Log            *logrus.Logger
}

func NewSekolahController(sekolahUseCase usecase.SekolahUseCase, log *logrus.Logger) SekolahController {
	return &sekolahControllerImpl{
		SekolahUseCase: sekolahUseCase,
		Log:            log,
	}
}

func (c *sekolahControllerImpl) Create(ctx fiber.Ctx) error {
	req := new(model.SekolahCreateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	sekolah, err := c.SekolahUseCase.Create(ctx.Context(), req)
	if err != nil {
		c.Log.Warnf("Failed create sekolah: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.SekolahResponse]{Data: sekolah})
}

func (c *sekolahControllerImpl) FindById(ctx fiber.Ctx) error {
	sekolahId := ctx.Params("id")

	sekolah, err := c.SekolahUseCase.FindById(ctx.Context(), sekolahId)
	if err != nil {
		c.Log.Warnf("Failed get sekolah by id: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.SekolahDetailResponse]{Data: sekolah})
}

func (c *sekolahControllerImpl) FindAll(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	sekolahs, _, err := c.SekolahUseCase.FindAll(ctx.Context(), search, page, size)
	if err != nil {
		c.Log.Warnf("Failed get all sekolah: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[[]*model.SekolahResponse]{
		Data: sekolahs,
		Paging: &model.PageMetaData{
			Page: page,
			Size: size,
		},
	})
}

func (c *sekolahControllerImpl) Update(ctx fiber.Ctx) error {
	req := new(model.SekolahUpdateRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return fiber.ErrBadRequest
	}

	sekolahId := ctx.Params("id")

	sekolah, err := c.SekolahUseCase.Update(ctx.Context(), sekolahId, req)
	if err != nil {
		c.Log.Warnf("Failed update sekolah: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.SekolahResponse]{Data: sekolah})
}

func (c *sekolahControllerImpl) Delete(ctx fiber.Ctx) error {
	sekolahId := ctx.Params("id")

	err := c.SekolahUseCase.Delete(ctx.Context(), sekolahId)
	if err != nil {
		c.Log.Warnf("Failed delete sekolah: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[string]{Data: "Sekolah deleted successfully"})
}
