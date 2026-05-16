package http

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
)

type ProgressController interface {
	Start(ctx fiber.Ctx) error
	Answer(ctx fiber.Ctx) error
	Scene(ctx fiber.Ctx) error
	Solve(ctx fiber.Ctx) error
	Submit(ctx fiber.Ctx) error
	GetSession(ctx fiber.Ctx) error
}

type progressControllerImpl struct {
	UseCase usecase.ProgressSessionUseCase
	Log     *logrus.Logger
}

func NewProgressController(uc usecase.ProgressSessionUseCase, log *logrus.Logger) ProgressController {
	return &progressControllerImpl{UseCase: uc, Log: log}
}

func (c *progressControllerImpl) Start(ctx fiber.Ctx) error {
	req := new(model.ProgressStartRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Start: bind error: %v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.Start(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Start: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusCreated).JSON(model.WebResponse[*model.ProgressStartResponse]{Data: resp})
}

func (c *progressControllerImpl) Answer(ctx fiber.Ctx) error {
	req := new(model.ProgressAnswerRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Answer: bind error: %v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.Answer(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Answer: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ProgressAnswerResponse]{Data: resp})
}

func (c *progressControllerImpl) Scene(ctx fiber.Ctx) error {
	req := new(model.ProgressSceneRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Scene: bind error: %v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.Scene(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Scene: %v", err)
		return err
	}

	// resp is nil when the scene is not an ending — return 200 with null data.
	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ProgressFinalizeResponse]{Data: resp})
}

func (c *progressControllerImpl) Solve(ctx fiber.Ctx) error {
	req := new(model.ProgressSolveRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Solve: bind error: %v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.Solve(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Solve: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ProgressFinalizeResponse]{Data: resp})
}

func (c *progressControllerImpl) Submit(ctx fiber.Ctx) error {
	req := new(model.ProgressSubmitRequest)
	if err := ctx.Bind().Body(req); err != nil {
		c.Log.Warnf("Submit: bind error: %v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.Submit(ctx.Context(), req, claims)
	if err != nil {
		c.Log.Warnf("Submit: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ProgressFinalizeResponse]{Data: resp})
}

func (c *progressControllerImpl) GetSession(ctx fiber.Ctx) error {
	contentType := ctx.Params("content_type")
	contentId, err := strconv.Atoi(ctx.Params("content_id"))
	if err != nil || contentId <= 0 {
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	resp, err := c.UseCase.GetSession(ctx.Context(), contentType, contentId, claims)
	if err != nil {
		c.Log.Warnf("GetSession: %v", err)
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON(model.WebResponse[*model.ProgressSessionResponse]{Data: resp})
}
