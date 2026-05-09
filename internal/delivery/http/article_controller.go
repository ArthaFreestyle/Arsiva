package http

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

type ArticleController interface {
	GetAllArticle(ctx fiber.Ctx) error
	GetArticleBySlug(ctx fiber.Ctx) error
	GetArticleById(ctx fiber.Ctx) error
	CreateArticle(ctx fiber.Ctx) error
	UpdateArticle(ctx fiber.Ctx) error
	DeleteArticle(ctx fiber.Ctx) error

	GetAllArticleManage(ctx fiber.Ctx) error
	GetArticleByIdManage(ctx fiber.Ctx) error
}

type articleControllerImpl struct {
	ArticleUseCase usecase.ArticleUseCase
	Log            *logrus.Logger
}

func NewArticleController(articleUseCase usecase.ArticleUseCase, log *logrus.Logger) ArticleController {
	return &articleControllerImpl{
		ArticleUseCase: articleUseCase,
		Log:            log,
	}
}

func (c *articleControllerImpl) GetAllArticle(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	articles, total, err := c.ArticleUseCase.GetAllArticle(ctx, page, size, search)
	if err != nil {
		c.Log.Warnf("Failed get all article : %+v", err)
		return err
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.ArticleResponse]{
		Data: articles,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetArticleBySlug(ctx fiber.Ctx) error {
	slug := ctx.Params("slug")
	article, err := c.ArticleUseCase.GetArticleBySlug(ctx, slug)
	if err != nil {
		c.Log.Warnf("Failed get article by slug : %+v", err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: article,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetArticleById(ctx fiber.Ctx) error {
	articleId := ctx.Params("id")
	article, err := c.ArticleUseCase.GetArticleById(ctx, articleId)
	if err != nil {
		c.Log.Warnf("Failed get article by id : %+v", err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: article,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) CreateArticle(ctx fiber.Ctx) error {
	article := new(model.ArticleCreateRequest)
	if err := ctx.Bind().Body(article); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}

	userId := ctx.Locals("userId").(string)
	createdArticle, err := c.ArticleUseCase.CreateArticle(ctx, article, userId)
	if err != nil {
		c.Log.Warnf("Failed create article : %+v", err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: createdArticle,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) UpdateArticle(ctx fiber.Ctx) error {
	articleId := ctx.Params("id")
	article := new(model.ArticleUpdateRequest)
	if err := ctx.Bind().Body(article); err != nil {
		c.Log.Warnf("Invalid request body : %+v", err)
		return fiber.ErrBadRequest
	}

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	updatedArticle, err := c.ArticleUseCase.UpdateArticleManage(ctx, article, articleId, userId, role)
	if err != nil {
		c.Log.Warnf("Failed update article : %+v", err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: updatedArticle,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) DeleteArticle(ctx fiber.Ctx) error {
	articleId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	err := c.ArticleUseCase.DeleteArticleManage(ctx, articleId, userId, role)
	if err != nil {
		c.Log.Warnf("Failed delete article : %+v", err)
		return err
	}

	res := model.WebResponse[any]{
		Data: "Article Deleted",
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetAllArticleManage(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	size, _ := strconv.Atoi(ctx.Query("size", "10"))
	search := ctx.Query("search", "")

	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	articles, total, err := c.ArticleUseCase.GetAllArticleManage(ctx, page, size, search, userId, role)
	if err != nil {
		c.Log.Warnf("Failed get all article manage : %+v", err)
		return err
	}

	totalPages := (total + size - 1) / size

	res := model.WebResponse[[]*model.ArticleResponse]{
		Data: articles,
		Paging: &model.PageMetaData{
			Page: page,
			Size: totalPages,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *articleControllerImpl) GetArticleByIdManage(ctx fiber.Ctx) error {
	articleId := ctx.Params("id")
	claims := ctx.Locals("user").(*model.Claims)
	userId := ctx.Locals("userId").(string)
	role := claims.Role

	article, err := c.ArticleUseCase.GetArticleByIdManage(ctx, articleId, userId, role)
	if err != nil {
		c.Log.Warnf("Failed get article by id manage : %+v", err)
		return err
	}

	res := model.WebResponse[*model.ArticleResponse]{
		Data: article,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
