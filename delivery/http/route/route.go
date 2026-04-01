package route

import (
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
	"ArthaFreestyle/Arsiva/delivery/http/middleware"
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
	CeritaController http.CeritaController
	StoryCategoryController http.StoryCategoryController
	QuizCategoryController http.QuizCategoryController
	AuthMiddleware fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/v1/login",c.AuthController.Login)
	c.App.Get("uploads/*",c.UploadController.GetFile)
}

func (c *RouteConfig) SetupAuthRoutes() {
	// Group with auth middleware
	auth := c.App.Group("/v1", c.AuthMiddleware)

	// ==========================================
	// SUPERADMIN ONLY
	// ==========================================
	superadmin := auth.Group("", middleware.RoleMiddleware("super_admin"))

	//users
	superadmin.Get("/users",c.UserController.GetAllUsers)
	superadmin.Get("/users/:id",c.UserController.GetUserById)
	superadmin.Post("/users",c.UserController.CreateUser)
	superadmin.Put("/users/:id",c.UserController.UpdateUser)
	superadmin.Delete("/users/:id",c.UserController.DeleteUser)

	// ==========================================
	// ALL AUTHENTICATED (member, guru, superadmin)
	// ==========================================
	allAuth := auth.Group("", middleware.RoleMiddleware("member","guru","super_admin"))

	//article category - read
	allAuth.Get("/categories/article",c.ArticleCategoryController.GetAllArticleCategories)
	allAuth.Get("/categories/article/:id",c.ArticleCategoryController.GetArticleCategoryById)

	//article - read
	allAuth.Get("/articles",c.ArticleController.GetAllArticle)
	allAuth.Get("/articles/detail/:id",c.ArticleController.GetArticleById)
	allAuth.Get("/articles/:slug",c.ArticleController.GetArticleBySlug)

	//puzzle - read
	allAuth.Get("/puzzles",c.PuzzleController.GetAllPuzzle)
	allAuth.Get("/puzzles/:id",c.PuzzleController.GetPuzzleById)

	//quiz - read
	allAuth.Get("/quizzes",c.QuizController.GetAllQuiz)
	allAuth.Get("/quizzes/:id",c.QuizController.GetQuizById)

	//cerita interaktif - read
	allAuth.Get("/stories",c.CeritaController.GetAllCerita)
	allAuth.Get("/stories/:id",c.CeritaController.GetCeritaById)

	//story category - read
	allAuth.Get("/categories/story",c.StoryCategoryController.GetAllStoryCategories)
	allAuth.Get("/categories/story/:id",c.StoryCategoryController.GetStoryCategoryById)

	//quiz category - read
	allAuth.Get("/categories/quiz",c.QuizCategoryController.GetAllQuizCategories)
	allAuth.Get("/categories/quiz/:id",c.QuizCategoryController.GetQuizCategoryById)

	// ==========================================
	// GURU + SUPERADMIN (content management)
	// ==========================================
	guruAdmin := auth.Group("", middleware.RoleMiddleware("guru","super_admin"))

	//article category - write
	guruAdmin.Post("/categories/article",c.ArticleCategoryController.CreateArticleCategory)
	guruAdmin.Put("/categories/article/:id",c.ArticleCategoryController.UpdateArticleCategory)
	guruAdmin.Delete("/categories/article/:id",c.ArticleCategoryController.DeleteArticleCategory)

	//article - write
	guruAdmin.Post("/articles",c.ArticleController.CreateArticle)
	guruAdmin.Put("/articles/:id",c.ArticleController.UpdateArticle)
	guruAdmin.Delete("/articles/:id",c.ArticleController.DeleteArticle)

	//puzzle - write
	guruAdmin.Post("/puzzles",c.PuzzleController.CreatePuzzle)
	guruAdmin.Put("/puzzles/:id",c.PuzzleController.UpdatePuzzle)
	guruAdmin.Delete("/puzzles/:id",c.PuzzleController.DeletePuzzle)

	//quiz - write
	guruAdmin.Post("/quizzes",c.QuizController.CreateQuiz)
	guruAdmin.Put("/quizzes/:id",c.QuizController.UpdateQuiz)
	guruAdmin.Delete("/quizzes/:id",c.QuizController.DeleteQuiz)

	//cerita interaktif - write
	guruAdmin.Post("/stories",c.CeritaController.CreateCerita)
	guruAdmin.Put("/stories/:id",c.CeritaController.UpdateCerita)
	guruAdmin.Delete("/stories/:id",c.CeritaController.DeleteCerita)

	//story category - write
	guruAdmin.Post("/categories/story",c.StoryCategoryController.CreateStoryCategory)
	guruAdmin.Put("/categories/story/:id",c.StoryCategoryController.UpdateStoryCategory)
	guruAdmin.Delete("/categories/story/:id",c.StoryCategoryController.DeleteStoryCategory)

	//quiz category - write
	guruAdmin.Post("/categories/quiz",c.QuizCategoryController.CreateQuizCategory)
	guruAdmin.Put("/categories/quiz/:id",c.QuizCategoryController.UpdateQuizCategory)
	guruAdmin.Delete("/categories/quiz/:id",c.QuizCategoryController.DeleteQuizCategory)

	//upload
	guruAdmin.Post("/upload/image",c.UploadController.UploadImage)
}