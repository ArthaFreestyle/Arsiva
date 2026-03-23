package route

import (
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
)

type RouteConfig struct {
	App *fiber.App
	AuthController http.AuthController
	UserController http.UserController
	ArticleCategoryController http.ArticleCategoryController
	ArticleController http.ArticleController
	UploadController http.UploadController
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/api/v1/login",c.AuthController.Login)
}

func (c *RouteConfig) SetupAuthRoutes() {
	//users
	c.App.Get("/api/v1/users",c.UserController.GetAllUsers)
	c.App.Get("/api/v1/users/:id",c.UserController.GetUserById)
	c.App.Post("/api/v1/users",c.UserController.CreateUser)
	c.App.Put("/api/v1/users/:id",c.UserController.UpdateUser)
	c.App.Delete("/api/v1/users/:id",c.UserController.DeleteUser)

	//article category
	c.App.Get("/api/v1/categories/article",c.ArticleCategoryController.GetAllArticleCategories)
	c.App.Get("/api/v1/categories/article/:id",c.ArticleCategoryController.GetArticleCategoryById)
	c.App.Post("/api/v1/categories/article",c.ArticleCategoryController.CreateArticleCategory)
	c.App.Put("/api/v1/categories/article/:id",c.ArticleCategoryController.UpdateArticleCategory)
	c.App.Delete("/api/v1/categories/article/:id",c.ArticleCategoryController.DeleteArticleCategory)

	//article
	c.App.Get("/api/v1/articles",c.ArticleController.GetAllArticle)
	c.App.Get("/api/v1/articles/detail/:id",c.ArticleController.GetArticleById)
	c.App.Get("/api/v1/articles/:slug",c.ArticleController.GetArticleBySlug)
	c.App.Post("/api/v1/articles",c.ArticleController.CreateArticle)
	c.App.Put("/api/v1/articles/:id",c.ArticleController.UpdateArticle)
	c.App.Delete("/api/v1/articles/:id",c.ArticleController.DeleteArticle)

	//upload
	c.App.Post("/api/v1/upload/image",c.UploadController.UploadImage)
}