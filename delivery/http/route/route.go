package route

import (
	"ArthaFreestyle/Arsiva/delivery/http/middleware"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
	"github.com/gofiber/fiber/v3"
)

type RouteConfig struct {
	App                       *fiber.App
	AuthController            http.AuthController
	UserController            http.UserController
	ArticleCategoryController http.ArticleCategoryController
	ArticleController         http.ArticleController
	UploadController          http.UploadController
	PuzzleController          http.PuzzleController
	QuizController            http.QuizController
	CeritaController          http.CeritaController
	StoryCategoryController   http.StoryCategoryController
	QuizCategoryController    http.QuizCategoryController
	GroupController           http.GroupController
	AuthMiddleware            fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/v1/login", c.AuthController.Login)
	c.App.Get("uploads/*", c.UploadController.GetFile)
}

func (c *RouteConfig) SetupAuthRoutes() {
	// Auth group — hanya prefix + auth middleware
	auth := c.App.Group("/v1", c.AuthMiddleware)

	// Role middleware disimpan dalam variabel
	superadminOnly := middleware.RoleMiddleware("super_admin")
	allRoles := middleware.RoleMiddleware("member", "guru", "super_admin")
	guruAdmin := middleware.RoleMiddleware("guru", "super_admin")
	guruOnly := middleware.RoleMiddleware("guru")
	memberOnly := middleware.RoleMiddleware("member")

	// ==========================================
	// SUPERADMIN ONLY
	// ==========================================

	// users
	auth.Get("/users", superadminOnly, c.UserController.GetAllUsers)
	auth.Get("/users/:id", superadminOnly, c.UserController.GetUserById)
	auth.Post("/users", superadminOnly, c.UserController.CreateUser)
	auth.Put("/users/:id", superadminOnly, c.UserController.UpdateUser)
	auth.Delete("/users/:id", superadminOnly, c.UserController.DeleteUser)

	// ==========================================
	// ALL AUTHENTICATED (member, guru, superadmin)
	// ==========================================

	// article category - read
	auth.Get("/categories/article", allRoles, c.ArticleCategoryController.GetAllArticleCategories)
	auth.Get("/categories/article/:id", allRoles, c.ArticleCategoryController.GetArticleCategoryById)

	// article - read
	auth.Get("/articles", allRoles, c.ArticleController.GetAllArticle)
	auth.Get("/articles/detail/:id", allRoles, c.ArticleController.GetArticleById)
	auth.Get("/articles/:slug", allRoles, c.ArticleController.GetArticleBySlug)

	// puzzle - read
	auth.Get("/puzzles", allRoles, c.PuzzleController.GetAllPuzzle)
	auth.Get("/puzzles/:id", allRoles, c.PuzzleController.GetPuzzleById)

	// quiz - read
	auth.Get("/quizzes", allRoles, c.QuizController.GetAllQuiz)
	auth.Get("/quizzes/:id", allRoles, c.QuizController.GetQuizById)

	// cerita interaktif - read
	auth.Get("/stories", allRoles, c.CeritaController.GetAllCerita)
	auth.Get("/stories/:id", allRoles, c.CeritaController.GetCeritaById)

	// story category - read
	auth.Get("/categories/story", allRoles, c.StoryCategoryController.GetAllStoryCategories)
	auth.Get("/categories/story/:id", allRoles, c.StoryCategoryController.GetStoryCategoryById)

	// quiz category - read
	auth.Get("/categories/quiz", allRoles, c.QuizCategoryController.GetAllQuizCategories)
	auth.Get("/categories/quiz/:id", allRoles, c.QuizCategoryController.GetQuizCategoryById)

	// ==========================================
	// GURU + SUPERADMIN (content management)
	// ==========================================

	// management READ endpoints (role-aware)
	auth.Get("/manage/puzzles", guruAdmin, c.PuzzleController.GetAllPuzzleManage)
	auth.Get("/manage/puzzles/:id", guruAdmin, c.PuzzleController.GetPuzzleByIdManage)
	auth.Get("/manage/quizzes", guruAdmin, c.QuizController.GetAllQuizManage)
	auth.Get("/manage/quizzes/:id", guruAdmin, c.QuizController.GetQuizByIdManage)
	auth.Get("/manage/stories", guruAdmin, c.CeritaController.GetAllCeritaManage)
	auth.Get("/manage/stories/:id", guruAdmin, c.CeritaController.GetCeritaByIdManage)

	// article category - write
	auth.Post("/categories/article", guruAdmin, c.ArticleCategoryController.CreateArticleCategory)
	auth.Put("/categories/article/:id", guruAdmin, c.ArticleCategoryController.UpdateArticleCategory)
	auth.Delete("/categories/article/:id", guruAdmin, c.ArticleCategoryController.DeleteArticleCategory)

	// article - write
	auth.Post("/articles", guruAdmin, c.ArticleController.CreateArticle)
	auth.Put("/articles/:id", guruAdmin, c.ArticleController.UpdateArticle)
	auth.Delete("/articles/:id", guruAdmin, c.ArticleController.DeleteArticle)

	// puzzle - write
	auth.Post("/puzzles", guruAdmin, c.PuzzleController.CreatePuzzle)
	auth.Put("/puzzles/:id", guruAdmin, c.PuzzleController.UpdatePuzzle)
	auth.Delete("/puzzles/:id", guruAdmin, c.PuzzleController.DeletePuzzle)

	// quiz - write
	auth.Post("/quizzes", guruAdmin, c.QuizController.CreateQuiz)
	auth.Put("/quizzes/:id", guruAdmin, c.QuizController.UpdateQuiz)
	auth.Delete("/quizzes/:id", guruAdmin, c.QuizController.DeleteQuiz)

	// cerita interaktif - write
	auth.Post("/stories", guruAdmin, c.CeritaController.CreateCerita)
	auth.Put("/stories/:id", guruAdmin, c.CeritaController.UpdateCerita)
	auth.Post("/stories/:id/scenes", guruAdmin, c.CeritaController.CreateScene)
	auth.Put("/stories/:id/scenes/:scene_id", guruAdmin, c.CeritaController.UpdateScene)
	auth.Delete("/stories/:id/scenes/:scene_id", guruAdmin, c.CeritaController.DeleteScene)
	auth.Delete("/stories/:id", guruAdmin, c.CeritaController.DeleteCerita)

	// story category - write
	auth.Post("/categories/story", guruAdmin, c.StoryCategoryController.CreateStoryCategory)
	auth.Put("/categories/story/:id", guruAdmin, c.StoryCategoryController.UpdateStoryCategory)
	auth.Delete("/categories/story/:id", guruAdmin, c.StoryCategoryController.DeleteStoryCategory)

	// quiz category - write
	auth.Post("/categories/quiz", guruAdmin, c.QuizCategoryController.CreateQuizCategory)
	auth.Put("/categories/quiz/:id", guruAdmin, c.QuizCategoryController.UpdateQuizCategory)
	auth.Delete("/categories/quiz/:id", guruAdmin, c.QuizCategoryController.DeleteQuizCategory)

	// upload
	auth.Post("/upload/image", guruAdmin, c.UploadController.UploadImage)

	// ==========================================
	// GURU ONLY (group management)
	// ==========================================

	// Group CRUD
	auth.Post("/groups", guruOnly, c.GroupController.CreateGroup)
	auth.Get("/groups", guruOnly, c.GroupController.GetAllGroups)
	auth.Get("/groups/:id", guruOnly, c.GroupController.GetGroupDetail)
	auth.Put("/groups/:id", guruOnly, c.GroupController.UpdateGroup)
	auth.Delete("/groups/:id", guruOnly, c.GroupController.DeleteGroup)

	// Group Member Management (Guru)
	auth.Post("/groups/:id/invite", guruOnly, c.GroupController.InviteMembersByEmail)
	auth.Get("/groups/:id/invite-link", guruOnly, c.GroupController.GenerateInviteLink)
	auth.Get("/groups/:id/members", guruOnly, c.GroupController.GetGroupMembers)
	auth.Delete("/groups/:id/members/:member_id", guruOnly, c.GroupController.RemoveMember)

	// ==========================================
	// MEMBER ONLY (join group)
	// ==========================================
	auth.Post("/groups/join", memberOnly, c.GroupController.JoinGroup)
}
