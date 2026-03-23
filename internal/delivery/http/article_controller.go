package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type ArticleController interface {
	GetAllArticle(ctx fiber.Ctx) (error)
	GetArticleBySlug(ctx fiber.Ctx) (error)
	GetArticleById(ctx fiber.Ctx) (error)
	CreateArticle(ctx fiber.Ctx) (error)
	UpdateArticle(ctx fiber.Ctx) (error)
	DeleteArticle(ctx fiber.Ctx) (error)
}

type articleControllerImpl struct {
	ArticleUseCase usecase.ArticleUseCase
	Log *logrus.Logger
}

func NewArticleController(articleUseCase usecase.ArticleUseCase,log *logrus.Logger) ArticleController {
	return &articleControllerImpl{
		ArticleUseCase: articleUseCase,
		Log: log,
	}
}

func (c *articleControllerImpl) GetAllArticle(ctx fiber.Ctx) (error) {
	articles,err := c.ArticleUseCase.GetAllArticle(ctx)
	if err != nil {
		c.Log.Warnf("Failed get all article : %+v",err)
		return err
	}

	res := model.WebResponse[[]*model.ArticleResponse]{
		Data: articles,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetArticleBySlug(ctx fiber.Ctx) (error) {
	slug := ctx.Params("slug")
	article,err := c.ArticleUseCase.GetArticleBySlug(ctx,slug)
	if err != nil {
		c.Log.Warnf("Failed get article by slug : %+v",err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: article,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetArticleById(ctx fiber.Ctx) (error) {
	articleId := ctx.Params("id")
	article,err := c.ArticleUseCase.GetArticleById(ctx,articleId)
	if err != nil {
		c.Log.Warnf("Failed get article by id : %+v",err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: article,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) CreateArticle(ctx fiber.Ctx) (error) {
	article := new(model.ArticleCreateRequest)
	if err := ctx.Bind().Body(article); err != nil {
		c.Log.Warnf("Invalid request body : %+v",err)
		return fiber.ErrBadRequest
	}

	userId := ctx.Locals("user_id").(string)
	createdArticle,err := c.ArticleUseCase.CreateArticle(ctx,article,userId)
	if err != nil {
		c.Log.Warnf("Failed create article : %+v",err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: createdArticle,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) UpdateArticle(ctx fiber.Ctx) (error) {
	articleId := ctx.Params("id")
	article := new(model.ArticleUpdateRequest)
	if err := ctx.Bind().Body(article); err != nil {
		c.Log.Warnf("Invalid request body : %+v",err)
		return fiber.ErrBadRequest
	}

	updatedArticle,err := c.ArticleUseCase.UpdateArticle(ctx,article,articleId)
	if err != nil {
		c.Log.Warnf("Failed update article : %+v",err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: updatedArticle,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) DeleteArticle(ctx fiber.Ctx) (error) {
	articleId := ctx.Params("id")
	err := c.ArticleUseCase.DeleteArticle(ctx,articleId)
	if err != nil {
		c.Log.Warnf("Failed delete article : %+v",err)
		return err
	}

	res := model.WebResponse[any]{
		Data: "Article Deleted",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
