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
	PuzzleController http.PuzzleController
	QuizController http.QuizController
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

	//puzzle
	c.App.Get("/api/v1/puzzles",c.PuzzleController.GetAllPuzzle)
	c.App.Get("/api/v1/puzzles/:id",c.PuzzleController.GetPuzzleById)
	c.App.Post("/api/v1/puzzles",c.PuzzleController.CreatePuzzle)
	c.App.Put("/api/v1/puzzles/:id",c.PuzzleController.UpdatePuzzle)
	c.App.Delete("/api/v1/puzzles/:id",c.PuzzleController.DeletePuzzle)

	//quiz
	c.App.Get("/api/v1/quizzes",c.QuizController.GetAllQuiz)
	c.App.Get("/api/v1/quizzes/:id",c.QuizController.GetQuizById)
	c.App.Post("/api/v1/quizzes",c.QuizController.CreateQuiz)
	c.App.Put("/api/v1/quizzes/:id",c.QuizController.UpdateQuiz)
	c.App.Delete("/api/v1/quizzes/:id",c.QuizController.DeleteQuiz)
	
	//upload
	c.App.Post("/api/v1/upload/image",c.UploadController.UploadImage)
	c.App.Get("/uploads/*",c.UploadController.GetFile)
}