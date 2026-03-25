package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type PuzzleController interface {
	GetAllPuzzle(ctx fiber.Ctx) (error)
	GetPuzzleById(ctx fiber.Ctx) (error)
	CreatePuzzle(ctx fiber.Ctx) (error)
	UpdatePuzzle(ctx fiber.Ctx) (error)
	DeletePuzzle(ctx fiber.Ctx) (error)
}

type puzzleControllerImpl struct {
	PuzzleUseCase usecase.PuzzleUseCase
	Log *logrus.Logger
}

func NewPuzzleController(puzzleUseCase usecase.PuzzleUseCase,log *logrus.Logger) PuzzleController {
	return &puzzleControllerImpl{
		PuzzleUseCase: puzzleUseCase,
		Log: log,
	}
}

func (c *puzzleControllerImpl) GetAllPuzzle(ctx fiber.Ctx) (error) {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	puzzles,total,err := c.PuzzleUseCase.GetAllPuzzle(ctx,page,size,search)
	if err != nil {
		c.Log.Warnf("error when get all puzzle: %v",err)
		return fiber.NewError(fiber.StatusInternalServerError,"internal server error")
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.PuzzleResponse]{
		Data: puzzles,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *puzzleControllerImpl) GetPuzzleById(ctx fiber.Ctx) (error) {
	puzzleId := ctx.Params("id")

	puzzle,err := c.PuzzleUseCase.GetPuzzleById(ctx,puzzleId)
	if err != nil {
		c.Log.Warnf("error when get puzzle by id: %v",err)
		return fiber.NewError(fiber.StatusInternalServerError,"internal server error")
	}

	res := model.WebResponse[*model.PuzzleResponse]{
		Data: puzzle,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *puzzleControllerImpl) CreatePuzzle(ctx fiber.Ctx) (error) {
	var puzzle model.PuzzleRequest
	if err := ctx.Bind().Body(&puzzle); err != nil {
		c.Log.Warnf("error when bind puzzle: %v",err)
		return fiber.NewError(fiber.StatusBadRequest,"bad request")
	}

	userId := ctx.Locals("userId").(string)

	createdPuzzle,err := c.PuzzleUseCase.CreatePuzzle(ctx,&puzzle,userId)
	if err != nil {
		c.Log.Warnf("error when create puzzle: %v",err)
		return fiber.NewError(fiber.StatusInternalServerError,"internal server error")
	}

	res := model.WebResponse[*model.PuzzleResponse]{
		Data: createdPuzzle,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *puzzleControllerImpl) UpdatePuzzle(ctx fiber.Ctx) (error) {
	var puzzle model.PuzzleRequest
	if err := ctx.Bind().Body(&puzzle); err != nil {
		c.Log.Warnf("error when bind puzzle: %v",err)
		return fiber.NewError(fiber.StatusBadRequest,"bad request")
	}

	puzzleId := ctx.Params("id")

	updatedPuzzle,err := c.PuzzleUseCase.UpdatePuzzle(ctx,&puzzle,puzzleId)
	if err != nil {
		c.Log.Warnf("error when update puzzle: %v",err)
		return fiber.NewError(fiber.StatusInternalServerError,"internal server error")
	}

	res := model.WebResponse[*model.PuzzleResponse]{
		Data: updatedPuzzle,
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *puzzleControllerImpl) DeletePuzzle(ctx fiber.Ctx) (error) {
	puzzleId := ctx.Params("id")

	err := c.PuzzleUseCase.DeletePuzzle(ctx,puzzleId)
	if err != nil {
		c.Log.Warnf("error when delete puzzle: %v",err)
		return fiber.NewError(fiber.StatusInternalServerError,"internal server error")
	}

	res := model.WebResponse[any]{
		Data: "puzzle deleted successfully",
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}