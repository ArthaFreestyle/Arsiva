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
	SekolahController         http.SekolahController
	GuruController            http.GuruController
	MemberController               http.MemberController
	MemberSocialLinkController     http.MemberSocialLinkController
	MemberAchievementController    http.MemberAchievementController
	AchievementController          http.AchievementController
	AuthMiddleware             fiber.Handler
	ProfileCompleteMiddleware  fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/v1/login", c.AuthController.Login)
	c.App.Post("/v1/register/member", c.AuthController.RegisterMember)
	c.App.Post("/v1/register/guru", c.AuthController.RegisterGuru)
	c.App.Get("uploads/*", c.UploadController.GetFile)
}

func (c *RouteConfig) SetupAuthRoutes() {
	// Auth group — prefix + auth middleware
	auth := c.App.Group("/v1", c.AuthMiddleware)

	// Role-only sub-groups (no profile-complete check).
	superadmin := auth.Group("", middleware.RoleMiddleware("super_admin"))
	allRoles   := auth.Group("", middleware.RoleMiddleware("member", "guru", "super_admin"))

	// Raw role sub-groups: role check only — used for profile creation / "me"
	// endpoints that must work BEFORE the profile is complete.
	guruRaw            := auth.Group("", middleware.RoleMiddleware("guru"))
	memberRaw          := auth.Group("", middleware.RoleMiddleware("member"))
	superadminOrGuru   := auth.Group("", middleware.RoleMiddleware("super_admin", "guru"))
	superadminOrMember := auth.Group("", middleware.RoleMiddleware("super_admin", "member"))

	// Action sub-groups: role check + profile-complete check. Half-onboarded
	// users cannot reach these endpoints.
	guruAdmin               := auth.Group("", middleware.RoleMiddleware("guru", "super_admin"), c.ProfileCompleteMiddleware)
	guruOnly                := auth.Group("", middleware.RoleMiddleware("guru"), c.ProfileCompleteMiddleware)
	memberOnly              := auth.Group("", middleware.RoleMiddleware("member"), c.ProfileCompleteMiddleware)
	superadminOrGuruReady   := auth.Group("", middleware.RoleMiddleware("super_admin", "guru"), c.ProfileCompleteMiddleware)
	superadminOrMemberReady := auth.Group("", middleware.RoleMiddleware("super_admin", "member"), c.ProfileCompleteMiddleware)

	// ==========================================
	// SUPERADMIN ONLY
	// ==========================================

	// users
	superadmin.Get("/users", c.UserController.GetAllUsers)
	guruOnly.Get("/users/search", c.UserController.SearchUsersByEmail)
	superadmin.Get("/users/:id", c.UserController.GetUserById)
	superadmin.Post("/users", c.UserController.CreateUser)
	superadmin.Put("/users/:id", c.UserController.UpdateUser)
	superadmin.Delete("/users/:id", c.UserController.DeleteUser)

	// ==========================================
	// ALL AUTHENTICATED (member, guru, superadmin)
	// ==========================================

	// article category - read
	allRoles.Get("/categories/article", c.ArticleCategoryController.GetAllArticleCategories)
	allRoles.Get("/categories/article/:id", c.ArticleCategoryController.GetArticleCategoryById)

	// article - read
	allRoles.Get("/articles", c.ArticleController.GetAllArticle)
	allRoles.Get("/articles/detail/:id", c.ArticleController.GetArticleById)
	allRoles.Get("/articles/:slug", c.ArticleController.GetArticleBySlug)

	// puzzle - read
	allRoles.Get("/puzzles", c.PuzzleController.GetAllPuzzle)
	allRoles.Get("/puzzles/:id", c.PuzzleController.GetPuzzleById)

	// quiz - read
	allRoles.Get("/quizzes", c.QuizController.GetAllQuiz)
	allRoles.Get("/quizzes/:id", c.QuizController.GetQuizById)

	// cerita interaktif - read
	allRoles.Get("/stories", c.CeritaController.GetAllCerita)
	allRoles.Get("/stories/:id", c.CeritaController.GetCeritaById)

	// story category - read
	allRoles.Get("/categories/story", c.StoryCategoryController.GetAllStoryCategories)
	allRoles.Get("/categories/story/:id", c.StoryCategoryController.GetStoryCategoryById)

	// quiz category - read
	allRoles.Get("/categories/quiz", c.QuizCategoryController.GetAllQuizCategories)
	allRoles.Get("/categories/quiz/:id", c.QuizCategoryController.GetQuizCategoryById)

	// ==========================================
	// GURU + SUPERADMIN (content management)
	// ==========================================

	// management READ endpoints (role-aware)
	guruAdmin.Get("/manage/articles", c.ArticleController.GetAllArticleManage)
	guruAdmin.Get("/manage/articles/:id", c.ArticleController.GetArticleByIdManage)
	guruAdmin.Get("/manage/puzzles", c.PuzzleController.GetAllPuzzleManage)
	guruAdmin.Get("/manage/puzzles/:id", c.PuzzleController.GetPuzzleByIdManage)
	guruAdmin.Get("/manage/quizzes", c.QuizController.GetAllQuizManage)
	guruAdmin.Get("/manage/quizzes/:id", c.QuizController.GetQuizByIdManage)
	guruAdmin.Get("/manage/stories", c.CeritaController.GetAllCeritaManage)
	guruAdmin.Get("/manage/stories/:id", c.CeritaController.GetCeritaByIdManage)

	// article category - write
	guruAdmin.Post("/categories/article", c.ArticleCategoryController.CreateArticleCategory)
	guruAdmin.Put("/categories/article/:id", c.ArticleCategoryController.UpdateArticleCategory)
	guruAdmin.Delete("/categories/article/:id", c.ArticleCategoryController.DeleteArticleCategory)

	// article - write
	guruAdmin.Post("/articles", c.ArticleController.CreateArticle)
	guruAdmin.Put("/articles/:id", c.ArticleController.UpdateArticle)
	guruAdmin.Delete("/articles/:id", c.ArticleController.DeleteArticle)

	// puzzle - write
	guruAdmin.Post("/puzzles", c.PuzzleController.CreatePuzzle)
	guruAdmin.Put("/puzzles/:id", c.PuzzleController.UpdatePuzzle)
	guruAdmin.Delete("/puzzles/:id", c.PuzzleController.DeletePuzzle)

	// quiz - write
	guruAdmin.Post("/quizzes", c.QuizController.CreateQuiz)
	guruAdmin.Put("/quizzes/:id", c.QuizController.UpdateQuiz)
	guruAdmin.Delete("/quizzes/:id", c.QuizController.DeleteQuiz)

	// cerita interaktif - write
	guruAdmin.Post("/stories", c.CeritaController.CreateCerita)
	guruAdmin.Put("/stories/:id", c.CeritaController.UpdateCerita)
	guruAdmin.Post("/stories/:id/scenes", c.CeritaController.CreateScene)
	guruAdmin.Put("/stories/:id/scenes/:scene_id", c.CeritaController.UpdateScene)
	guruAdmin.Delete("/stories/:id/scenes/:scene_id", c.CeritaController.DeleteScene)
	guruAdmin.Delete("/stories/:id", c.CeritaController.DeleteCerita)

	// story category - write
	guruAdmin.Post("/categories/story", c.StoryCategoryController.CreateStoryCategory)
	guruAdmin.Put("/categories/story/:id", c.StoryCategoryController.UpdateStoryCategory)
	guruAdmin.Delete("/categories/story/:id", c.StoryCategoryController.DeleteStoryCategory)

	// quiz category - write
	guruAdmin.Post("/categories/quiz", c.QuizCategoryController.CreateQuizCategory)
	guruAdmin.Put("/categories/quiz/:id", c.QuizCategoryController.UpdateQuizCategory)
	guruAdmin.Delete("/categories/quiz/:id", c.QuizCategoryController.DeleteQuizCategory)

	// upload
	guruAdmin.Post("/upload/image", c.UploadController.UploadImage)

	// ==========================================
	// GURU ONLY (group management)
	// ==========================================

	// Group CRUD
	guruOnly.Post("/groups", c.GroupController.CreateGroup)
	guruOnly.Get("/groups", c.GroupController.GetAllGroups)
	guruOnly.Get("/groups/:id", c.GroupController.GetGroupDetail)
	guruOnly.Put("/groups/:id", c.GroupController.UpdateGroup)
	guruOnly.Delete("/groups/:id", c.GroupController.DeleteGroup)

	// Group Member Management (Guru)
	guruOnly.Post("/groups/:id/invite", c.GroupController.InviteMembersByEmail)
	guruOnly.Get("/groups/:id/invite-link", c.GroupController.GenerateInviteLink)
	guruOnly.Get("/groups/:id/members", c.GroupController.GetGroupMembers)
	guruOnly.Delete("/groups/:id/members/:member_id", c.GroupController.RemoveMember)

	// Group Content Management
	guruOnly.Post("/groups/:id/contents", c.GroupController.AddContent)
	guruOnly.Delete("/groups/:id/contents/:content_id", c.GroupController.RemoveContent)

	// View group contents — usecase handles ownership (guru) vs membership (member) check
	allRoles.Get("/groups/:id/contents", c.GroupController.GetGroupContents)

	// ==========================================
	// MEMBER ONLY (join group)
	// ==========================================
	memberOnly.Post("/groups/join", c.GroupController.JoinGroup)

	// ==========================================
	// SEKOLAH
	//   - read: semua role ter-autentikasi
	//   - write: super_admin only
	// ==========================================
	allRoles.Get("/sekolah", c.SekolahController.FindAll)
	allRoles.Get("/sekolah/:id", c.SekolahController.FindById)
	superadmin.Post("/sekolah", c.SekolahController.Create)
	superadmin.Put("/sekolah/:id", c.SekolahController.Update)
	superadmin.Delete("/sekolah/:id", c.SekolahController.Delete)

	// ==========================================
	// GURU MANAGEMENT
	// ==========================================
	// POST /guru and GET /guru/me are allowed while profile is incomplete
	// (POST creates the profile, GET lets the FE detect "no profile yet").
	superadminOrGuru.Post("/guru", c.GuruController.Create)       // profile creation — no check
	superadmin.Get("/guru", c.GuruController.FindAll)
	guruRaw.Get("/guru/me", c.GuruController.GetMe)               // allowed while incomplete
	superadminOrGuruReady.Get("/guru/:id", c.GuruController.FindById)
	superadminOrGuruReady.Put("/guru/:id", c.GuruController.Update)
	superadmin.Delete("/guru/:id", c.GuruController.Delete)

	// ==========================================
	// MEMBER MANAGEMENT
	// ==========================================
	// POST /member and GET /member/me are allowed while profile is incomplete.
	superadminOrMember.Post("/member", c.MemberController.Create) // profile creation — no check
	superadmin.Get("/member", c.MemberController.FindAll)
	memberRaw.Get("/member/me", c.MemberController.GetMe)         // allowed while incomplete
	memberOnly.Put("/member/me", c.MemberController.UpdateMe)
	memberOnly.Get("/member/profile", c.MemberController.GetProfile)
	superadminOrMemberReady.Get("/member/:id", c.MemberController.FindById)
	superadminOrMemberReady.Put("/member/:id", c.MemberController.Update)
	superadmin.Delete("/member/:id", c.MemberController.Delete)

	// ==========================================
	// ACHIEVEMENTS
	//   - read: semua role ter-autentikasi
	//   - write: super_admin only
	// ==========================================
	allRoles.Get("/achievements", c.AchievementController.FindAll)
	allRoles.Get("/achievements/:id", c.AchievementController.FindById)
	superadmin.Post("/achievements", c.AchievementController.Create)
	superadmin.Put("/achievements/:id", c.AchievementController.Update)
	superadmin.Delete("/achievements/:id", c.AchievementController.Delete)

	// ==========================================
	// MEMBER SOCIAL LINKS (member only)
	// ==========================================
	memberOnly.Post("/member/social-links", c.MemberSocialLinkController.Create)
	memberOnly.Get("/member/social-links", c.MemberSocialLinkController.FindAllMine)
	memberOnly.Get("/member/social-links/:id", c.MemberSocialLinkController.FindById)
	memberOnly.Put("/member/social-links/:id", c.MemberSocialLinkController.Update)
	memberOnly.Delete("/member/social-links/:id", c.MemberSocialLinkController.Delete)

	// ==========================================
	// MEMBER ACHIEVEMENTS
	//   - create / read: member only (self)
	//   - delete: super_admin only (corrective)
	// ==========================================
	memberOnly.Post("/member/achievements", c.MemberAchievementController.Create)
	memberOnly.Get("/member/achievements", c.MemberAchievementController.FindAllMine)
	memberOnly.Get("/member/achievements/:achievement_id", c.MemberAchievementController.FindOne)
	superadmin.Delete("/member/achievements/:member_id/:achievement_id", c.MemberAchievementController.Delete)
}
