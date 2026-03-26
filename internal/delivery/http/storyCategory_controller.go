package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type StoryCategoryController interface {
	GetAllStoryCategories(ctx fiber.Ctx) error
	GetStoryCategoryById(ctx fiber.Ctx) error
	CreateStoryCategory(ctx fiber.Ctx) error
	UpdateStoryCategory(ctx fiber.Ctx) error
	DeleteStoryCategory(ctx fiber.Ctx) error
}

type storyCategoryControllerImpl struct {
	StoryCategoryUseCase usecase.StoryCategoryUseCase
	Log                  *logrus.Logger
}

func NewStoryCategoryController(storyCategoryUseCase usecase.StoryCategoryUseCase, log *logrus.Logger) StoryCategoryController {
	return &storyCategoryControllerImpl{
		StoryCategoryUseCase: storyCategoryUseCase,
		Log:                  log,
	}
}

func (c *storyCategoryControllerImpl) GetAllStoryCategories(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	categories, total, err := c.StoryCategoryUseCase.GetAllStoryCategories(ctx, page, size, search)
	if err != nil {
		c.Log.Warnf("Failed get all story categories : %+v", err)
		return err
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.StoryCategoryResponse]{
		Data: categories,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *storyCategoryControllerImpl) GetStoryCategoryById(ctx fiber.Ctx) error {
	storyCategoryId := ctx.Params("id")
	category, err := c.StoryCategoryUseCase.GetStoryCategoryById(ctx, storyCategoryId)
	if err != nil {
		c.Log.Warnf("Failed get story category by id : %+v", err)
		return err
	}

	res := model.WebResponse[*model.StoryCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *storyCategoryControllerImpl) CreateStoryCategory(ctx fiber.Ctx) error {
	storyCategory := new(model.StoryCategoryRequest)
	if err := ctx.Bind().Body(storyCategory); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}
	category, err := c.StoryCategoryUseCase.CreateStoryCategory(ctx, storyCategory)
	if err != nil {
		c.Log.Warnf("Failed create story category : %+v", err)
		return err
	}
	res := model.WebResponse[*model.StoryCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *storyCategoryControllerImpl) UpdateStoryCategory(ctx fiber.Ctx) error {
	storyCategoryId := ctx.Params("id")
	storyCategory := new(model.StoryCategoryRequest)
	if err := ctx.Bind().Body(storyCategory); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}
	category, err := c.StoryCategoryUseCase.UpdateStoryCategory(ctx, storyCategory, storyCategoryId)
	if err != nil {
		c.Log.Warnf("Failed update story category : %+v", err)
		return err
	}
	res := model.WebResponse[*model.StoryCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *storyCategoryControllerImpl) DeleteStoryCategory(ctx fiber.Ctx) error {
	storyCategoryId := ctx.Params("id")
	err := c.StoryCategoryUseCase.DeleteStoryCategory(ctx, storyCategoryId)
	if err != nil {
		c.Log.Warnf("Failed delete story category : %+v", err)
		return err
	}
	res := model.WebResponse[any]{
		Data: "Story Category Deleted",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
