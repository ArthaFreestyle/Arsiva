package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type QuizCategoryController interface {
	GetAllQuizCategories(ctx fiber.Ctx) error
	GetQuizCategoryById(ctx fiber.Ctx) error
	CreateQuizCategory(ctx fiber.Ctx) error
	UpdateQuizCategory(ctx fiber.Ctx) error
	DeleteQuizCategory(ctx fiber.Ctx) error
}

type quizCategoryControllerImpl struct {
	QuizCategoryUseCase usecase.QuizCategoryUseCase
	Log                 *logrus.Logger
}

func NewQuizCategoryController(quizCategoryUseCase usecase.QuizCategoryUseCase, log *logrus.Logger) QuizCategoryController {
	return &quizCategoryControllerImpl{
		QuizCategoryUseCase: quizCategoryUseCase,
		Log:                 log,
	}
}

func (c *quizCategoryControllerImpl) GetAllQuizCategories(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	categories, total, err := c.QuizCategoryUseCase.GetAllQuizCategories(ctx, page, size, search)
	if err != nil {
		c.Log.Warnf("Failed get all quiz categories : %+v", err)
		return err
	}

	totalPages := (total + size - 1) / size
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	res := model.WebResponse[[]*model.QuizCategoryResponse]{
		Data: categories,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizCategoryControllerImpl) GetQuizCategoryById(ctx fiber.Ctx) error {
	quizCategoryId := ctx.Params("id")
	category, err := c.QuizCategoryUseCase.GetQuizCategoryById(ctx, quizCategoryId)
	if err != nil {
		c.Log.Warnf("Failed get quiz category by id : %+v", err)
		return err
	}

	res := model.WebResponse[*model.QuizCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizCategoryControllerImpl) CreateQuizCategory(ctx fiber.Ctx) error {
	quizCategory := new(model.QuizCategoryRequest)
	if err := ctx.Bind().Body(quizCategory); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}
	userId := ctx.Locals("userId").(string)
	category, err := c.QuizCategoryUseCase.CreateQuizCategory(ctx, quizCategory, userId)
	if err != nil {
		c.Log.Warnf("Failed create quiz category : %+v", err)
		return err
	}
	res := model.WebResponse[*model.QuizCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizCategoryControllerImpl) UpdateQuizCategory(ctx fiber.Ctx) error {
	quizCategoryId := ctx.Params("id")
	quizCategory := new(model.QuizCategoryRequest)
	if err := ctx.Bind().Body(quizCategory); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}
	category, err := c.QuizCategoryUseCase.UpdateQuizCategory(ctx, quizCategory, quizCategoryId)
	if err != nil {
		c.Log.Warnf("Failed update quiz category : %+v", err)
		return err
	}
	res := model.WebResponse[*model.QuizCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizCategoryControllerImpl) DeleteQuizCategory(ctx fiber.Ctx) error {
	quizCategoryId := ctx.Params("id")
	err := c.QuizCategoryUseCase.DeleteQuizCategory(ctx, quizCategoryId)
	if err != nil {
		c.Log.Warnf("Failed delete quiz category : %+v", err)
		return err
	}
	res := model.WebResponse[any]{
		Data: "Quiz Category Deleted",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
