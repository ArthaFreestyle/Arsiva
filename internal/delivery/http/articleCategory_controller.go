package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"ArthaFreestyle/Arsiva/internal/model"
)

type ArticleCategoryController interface {
	GetAllArticleCategories(ctx fiber.Ctx) (error)
	GetArticleCategoryById(ctx fiber.Ctx) (error)
	CreateArticleCategory(ctx fiber.Ctx) (error)
	UpdateArticleCategory(ctx fiber.Ctx) (error)
	DeleteArticleCategory(ctx fiber.Ctx) (error)
}

type articleCategoryControllerImpl struct {
	ArticleCategoryUseCase usecase.ArticleCategoryUseCase
	Log *logrus.Logger
}

func NewArticleCategoryController(articleCategoryUseCase usecase.ArticleCategoryUseCase,log *logrus.Logger) ArticleCategoryController {
	return &articleCategoryControllerImpl{
		ArticleCategoryUseCase: articleCategoryUseCase,
		Log: log,
	}
}

func (c *articleCategoryControllerImpl) GetAllArticleCategories(ctx fiber.Ctx) (error) {
	category,err := c.ArticleCategoryUseCase.GetAllArticleCategories(ctx)
	if err != nil {
		return err
	}
	
	res := model.WebResponse[[]*model.ArticleCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleCategoryControllerImpl) GetArticleCategoryById(ctx fiber.Ctx) (error) {
	articleCategoryId := ctx.Params("id")
	category,err := c.ArticleCategoryUseCase.GetArticleCategoryById(ctx,articleCategoryId)
	if err != nil {
		return err
	}
	
	res := model.WebResponse[*model.ArticleCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleCategoryControllerImpl) CreateArticleCategory(ctx fiber.Ctx) (error) {
	articleCategory := new(model.ArticleCategoryRequest)
	if err := ctx.Bind().Body(articleCategory); err != nil {
		return fiber.ErrBadRequest
	}
	category,err := c.ArticleCategoryUseCase.CreateArticleCategory(ctx,articleCategory)
	if err != nil {
		return err
	}
	res := model.WebResponse[*model.ArticleCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleCategoryControllerImpl) UpdateArticleCategory(ctx fiber.Ctx) (error) {
	articleCategoryId := ctx.Params("id")
	articleCategory := new(model.ArticleCategoryRequest)
	if err := ctx.Bind().Body(articleCategory); err != nil {
		return fiber.ErrBadRequest
	}
	category,err := c.ArticleCategoryUseCase.UpdateArticleCategory(ctx,articleCategory,articleCategoryId)
	if err != nil {
		return err
	}
	res := model.WebResponse[*model.ArticleCategoryResponse]{
		Data: category,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleCategoryControllerImpl) DeleteArticleCategory(ctx fiber.Ctx) (error) {
	articleCategoryId := ctx.Params("id")
	err := c.ArticleCategoryUseCase.DeleteArticleCategory(ctx,articleCategoryId)
	if err != nil {
		return err
	}
	res := model.WebResponse[any]{
		Data: "Article Category Deleted",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}