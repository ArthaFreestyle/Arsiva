package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type QuizController interface {
	GetAllQuiz(ctx fiber.Ctx) error
	GetQuizById(ctx fiber.Ctx) error
	CreateQuiz(ctx fiber.Ctx) error
	UpdateQuiz(ctx fiber.Ctx) error
	DeleteQuiz(ctx fiber.Ctx) error
}

type quizControllerImpl struct {
	QuizUseCase usecase.QuizUseCase
	Log         *logrus.Logger
}

func NewQuizController(quizUseCase usecase.QuizUseCase, log *logrus.Logger) QuizController {
	return &quizControllerImpl{
		QuizUseCase: quizUseCase,
		Log:         log,
	}
}

func (c *quizControllerImpl) GetAllQuiz(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	quizzes, total, err := c.QuizUseCase.GetAll(ctx, page, size, search)
	if err != nil {
		c.Log.Warnf("error when get all quiz: %v", err)
		return err
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.QuizResponse]{
		Data: quizzes,
		Paging: &model.PageMetaData{
			Page:  page,
			Size:  totalPages,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizControllerImpl) GetQuizById(ctx fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid quiz id")
	}

	quiz, err := c.QuizUseCase.GetByID(ctx, id)
	if err != nil {
		c.Log.Warnf("error when get quiz by id: %v", err)
		return err
	}

	res := model.WebResponse[*model.QuizResponse]{
		Data: quiz,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizControllerImpl) CreateQuiz(ctx fiber.Ctx) error {
	var quiz model.QuizRequest
	if err := ctx.Bind().Body(&quiz); err != nil {
		c.Log.Warnf("error when bind quiz: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	userId := ctx.Locals("userId").(string)

	createdQuiz, err := c.QuizUseCase.Create(ctx, &quiz, userId)
	if err != nil {
		c.Log.Warnf("error when create quiz: %v", err)
		return err
	}

	res := model.WebResponse[*model.QuizResponse]{
		Data: createdQuiz,
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (c *quizControllerImpl) UpdateQuiz(ctx fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid quiz id")
	}

	var quiz model.QuizRequest
	if err := ctx.Bind().Body(&quiz); err != nil {
		c.Log.Warnf("error when bind quiz: %v", err)
		return fiber.NewError(fiber.StatusBadRequest, "bad request")
	}

	updatedQuiz, err := c.QuizUseCase.Update(ctx, &quiz, id)
	if err != nil {
		c.Log.Warnf("error when update quiz: %v", err)
		return err
	}

	res := model.WebResponse[*model.QuizResponse]{
		Data: updatedQuiz,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *quizControllerImpl) DeleteQuiz(ctx fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid quiz id")
	}

	err = c.QuizUseCase.Delete(ctx, id)
	if err != nil {
		c.Log.Warnf("error when delete quiz: %v", err)
		return err
	}

	res := model.WebResponse[any]{
		Data: "quiz deleted successfully",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}