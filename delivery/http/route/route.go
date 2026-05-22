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
	ProgressController             http.ProgressController
	LeaderboardController          http.LeaderboardController
	AuthMiddleware             fiber.Handler
	ProfileCompleteMiddleware  fiber.Handler
	AuthLimiter                fiber.Handler
}

func (c *RouteConfig) SetupRoutes() {
	c.SetupGuestRoutes()
	c.SetupAuthRoutes()
}

func (c *RouteConfig) SetupGuestRoutes() {
	c.App.Post("/v1/login", c.AuthLimiter, c.AuthController.Login)
	c.App.Post("/v1/register/member", c.AuthLimiter, c.AuthController.RegisterMember)
	c.App.Post("/v1/register/guru", c.AuthLimiter, c.AuthController.RegisterGuru)
	c.App.Get("uploads/*", c.UploadController.GetFile)
}

func (c *RouteConfig) SetupAuthRoutes() {
	// Auth group — prefix + auth middleware
	auth := c.App.Group("/v1", c.AuthMiddleware)

	// Single role middlewares (built once, reused).
	superadminOnly       := middleware.RoleMiddleware("super_admin")
	allRoles             := middleware.RoleMiddleware("member", "guru", "super_admin")
	guruRole             := middleware.RoleMiddleware("guru")
	memberRole           := middleware.RoleMiddleware("member")
	guruAdminRole        := middleware.RoleMiddleware("guru", "super_admin")
	superadminOrGuruRole := middleware.RoleMiddleware("super_admin", "guru")
	superadminOrMemberRole := middleware.RoleMiddleware("super_admin", "member")

	// Profile-complete middleware shorthand.
	pc := c.ProfileCompleteMiddleware

	// ==========================================
	// SUPERADMIN ONLY
	// ==========================================

	// users
	auth.Get("/users", superadminOnly, c.UserController.GetAllUsers)
	auth.Get("/users/search", guruRole, pc, c.UserController.SearchUsersByEmail)
	auth.Get("/users/deleted", superadminOnly, c.UserController.GetDeletedUsers)
	auth.Get("/users/:id", superadminOnly, c.UserController.GetUserById)
	auth.Post("/users", superadminOnly, c.UserController.CreateUser)
	auth.Put("/users/:id", superadminOnly, c.UserController.UpdateUser)
	auth.Delete("/users/:id", superadminOnly, c.UserController.DeleteUser)
	auth.Patch("/users/:id/restore", superadminOnly, c.UserController.RestoreUser)

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
	auth.Get("/manage/articles", guruAdminRole, pc, c.ArticleController.GetAllArticleManage)
	auth.Get("/manage/articles/:id", guruAdminRole, pc, c.ArticleController.GetArticleByIdManage)
	auth.Get("/manage/puzzles", guruAdminRole, pc, c.PuzzleController.GetAllPuzzleManage)
	auth.Get("/manage/puzzles/:id", guruAdminRole, pc, c.PuzzleController.GetPuzzleByIdManage)
	auth.Get("/manage/quizzes", guruAdminRole, pc, c.QuizController.GetAllQuizManage)
	auth.Get("/manage/quizzes/:id", guruAdminRole, pc, c.QuizController.GetQuizByIdManage)
	auth.Get("/manage/stories", guruAdminRole, pc, c.CeritaController.GetAllCeritaManage)
	auth.Get("/manage/stories/:id", guruAdminRole, pc, c.CeritaController.GetCeritaByIdManage)

	// article category - write
	auth.Post("/categories/article", guruAdminRole, pc, c.ArticleCategoryController.CreateArticleCategory)
	auth.Put("/categories/article/:id", guruAdminRole, pc, c.ArticleCategoryController.UpdateArticleCategory)
	auth.Delete("/categories/article/:id", guruAdminRole, pc, c.ArticleCategoryController.DeleteArticleCategory)

	// article - write
	auth.Post("/articles", guruAdminRole, pc, c.ArticleController.CreateArticle)
	auth.Put("/articles/:id", guruAdminRole, pc, c.ArticleController.UpdateArticle)
	auth.Delete("/articles/:id", guruAdminRole, pc, c.ArticleController.DeleteArticle)

	// puzzle - write
	auth.Post("/puzzles", guruAdminRole, pc, c.PuzzleController.CreatePuzzle)
	auth.Put("/puzzles/:id", guruAdminRole, pc, c.PuzzleController.UpdatePuzzle)
	auth.Delete("/puzzles/:id", guruAdminRole, pc, c.PuzzleController.DeletePuzzle)

	// quiz - write
	auth.Post("/quizzes", guruAdminRole, pc, c.QuizController.CreateQuiz)
	auth.Put("/quizzes/:id", guruAdminRole, pc, c.QuizController.UpdateQuiz)
	auth.Delete("/quizzes/:id", guruAdminRole, pc, c.QuizController.DeleteQuiz)

	// cerita interaktif - write
	auth.Post("/stories", guruAdminRole, pc, c.CeritaController.CreateCerita)
	auth.Put("/stories/:id", guruAdminRole, pc, c.CeritaController.UpdateCerita)
	auth.Post("/stories/:id/scenes", guruAdminRole, pc, c.CeritaController.CreateScene)
	auth.Put("/stories/:id/scenes/:scene_id", guruAdminRole, pc, c.CeritaController.UpdateScene)
	auth.Delete("/stories/:id/scenes/:scene_id", guruAdminRole, pc, c.CeritaController.DeleteScene)
	auth.Delete("/stories/:id", guruAdminRole, pc, c.CeritaController.DeleteCerita)

	// story category - write
	auth.Post("/categories/story", guruAdminRole, pc, c.StoryCategoryController.CreateStoryCategory)
	auth.Put("/categories/story/:id", guruAdminRole, pc, c.StoryCategoryController.UpdateStoryCategory)
	auth.Delete("/categories/story/:id", guruAdminRole, pc, c.StoryCategoryController.DeleteStoryCategory)

	// quiz category - write
	auth.Post("/categories/quiz", guruAdminRole, pc, c.QuizCategoryController.CreateQuizCategory)
	auth.Put("/categories/quiz/:id", guruAdminRole, pc, c.QuizCategoryController.UpdateQuizCategory)
	auth.Delete("/categories/quiz/:id", guruAdminRole, pc, c.QuizCategoryController.DeleteQuizCategory)

	// upload
	auth.Post("/upload/image", guruAdminRole, pc, c.UploadController.UploadImage)

	// ==========================================
	// GURU ONLY (group management)
	// ==========================================

	// Group CRUD
	auth.Post("/groups", guruRole, pc, c.GroupController.CreateGroup)
	auth.Get("/groups", guruRole, pc, c.GroupController.GetAllGroups)
	auth.Get("/groups/:id", guruRole, pc, c.GroupController.GetGroupDetail)
	auth.Put("/groups/:id", guruRole, pc, c.GroupController.UpdateGroup)
	auth.Delete("/groups/:id", guruRole, pc, c.GroupController.DeleteGroup)

	// Group Member Management (Guru)
	auth.Post("/groups/:id/invite", guruRole, pc, c.GroupController.InviteMembersByEmail)
	auth.Get("/groups/:id/invite-link", guruRole, pc, c.GroupController.GenerateInviteLink)
	auth.Get("/groups/:id/members", guruRole, pc, c.GroupController.GetGroupMembers)
	auth.Delete("/groups/:id/members/:member_id", guruRole, pc, c.GroupController.RemoveMember)

	// Group Content Management
	auth.Post("/groups/:id/contents", guruRole, pc, c.GroupController.AddContent)
	auth.Delete("/groups/:id/contents/:content_id", guruRole, pc, c.GroupController.RemoveContent)

	// View group contents — usecase handles ownership (guru) vs membership (member) check
	auth.Get("/groups/:id/contents", allRoles, c.GroupController.GetGroupContents)

	// ==========================================
	// MEMBER ONLY (join group)
	// ==========================================
	auth.Post("/groups/join", memberRole, pc, c.GroupController.JoinGroup)

	// ==========================================
	// SEKOLAH
	//   - read: semua role ter-autentikasi
	//   - write: super_admin only
	// ==========================================
	auth.Get("/sekolah", allRoles, c.SekolahController.FindAll)
	auth.Get("/sekolah/:id", allRoles, c.SekolahController.FindById)
	auth.Post("/sekolah", superadminOnly, c.SekolahController.Create)
	auth.Put("/sekolah/:id", superadminOnly, c.SekolahController.Update)
	auth.Delete("/sekolah/:id", superadminOnly, c.SekolahController.Delete)

	// ==========================================
	// GURU MANAGEMENT
	// ==========================================
	// POST /guru and GET /guru/me are allowed while profile is incomplete
	// (POST creates the profile, GET lets the FE detect "no profile yet").
	auth.Post("/guru", superadminOrGuruRole, c.GuruController.Create)       // profile creation — no pc
	auth.Get("/guru", superadminOnly, c.GuruController.FindAll)
	auth.Get("/guru/me", guruRole, c.GuruController.GetMe)                  // allowed while incomplete
	auth.Get("/guru/:id", superadminOrGuruRole, pc, c.GuruController.FindById)
	auth.Put("/guru/:id", superadminOrGuruRole, pc, c.GuruController.Update)
	auth.Delete("/guru/:id", superadminOnly, c.GuruController.Delete)

	// ==========================================
	// MEMBER MANAGEMENT
	// ==========================================
	// POST /member and GET /member/me are allowed while profile is incomplete.
	auth.Post("/member", superadminOrMemberRole, c.MemberController.Create) // profile creation — no pc
	auth.Get("/member", superadminOnly, c.MemberController.FindAll)
	auth.Get("/member/me", memberRole, c.MemberController.GetMe)            // allowed while incomplete
	auth.Put("/member/me", memberRole, pc, c.MemberController.UpdateMe)
	auth.Get("/member/profile", memberRole, pc, c.MemberController.GetProfile)
	// NOTE: the wildcard "/member/:id" routes are registered AFTER the static
	// "/member/social-links" and "/member/achievements" routes below. Fiber
	// matches routes in registration order, so a "/member/:id" registered here
	// would greedily capture "/member/achievements" (:id="achievements") and
	// shadow the static route, returning 403 from FindById's self-only check.

	// ==========================================
	// ACHIEVEMENTS
	//   - read: semua role ter-autentikasi
	//   - write: super_admin only
	// ==========================================
	auth.Get("/achievements", allRoles, c.AchievementController.FindAll)
	auth.Get("/achievements/:id", allRoles, c.AchievementController.FindById)
	auth.Post("/achievements", superadminOnly, c.AchievementController.Create)
	auth.Put("/achievements/:id", superadminOnly, c.AchievementController.Update)
	auth.Delete("/achievements/:id", superadminOnly, c.AchievementController.Delete)

	// ==========================================
	// MEMBER SOCIAL LINKS (member only)
	// ==========================================
	auth.Post("/member/social-links", memberRole, pc, c.MemberSocialLinkController.Create)
	auth.Get("/member/social-links", memberRole, pc, c.MemberSocialLinkController.FindAllMine)
	auth.Get("/member/social-links/:id", memberRole, pc, c.MemberSocialLinkController.FindById)
	auth.Put("/member/social-links/:id", memberRole, pc, c.MemberSocialLinkController.Update)
	auth.Delete("/member/social-links/:id", memberRole, pc, c.MemberSocialLinkController.Delete)

	// ==========================================
	// MEMBER ACHIEVEMENTS
	//   - create / read: member only (self)
	//   - delete: super_admin only (corrective)
	// ==========================================
	auth.Post("/member/achievements", memberRole, pc, c.MemberAchievementController.Create)
	auth.Get("/member/achievements", memberRole, pc, c.MemberAchievementController.FindAllMine)
	auth.Get("/member/achievements/:achievement_id", memberRole, pc, c.MemberAchievementController.FindOne)
	auth.Delete("/member/achievements/:member_id/:achievement_id", superadminOnly, c.MemberAchievementController.Delete)

	// ==========================================
	// MEMBER BY ID (wildcard — registered last so static "/member/*" routes
	// above take precedence under Fiber's registration-order matching)
	// ==========================================
	auth.Get("/member/:id", superadminOrMemberRole, pc, c.MemberController.FindById)
	auth.Put("/member/:id", superadminOrMemberRole, pc, c.MemberController.Update)
	auth.Delete("/member/:id", superadminOnly, c.MemberController.Delete)

	// ==========================================
	// LEADERBOARD
	//   - public: all authenticated roles, no profile-complete required
	//   - group:  all authenticated roles, profile-complete required
	// ==========================================
	auth.Get("/leaderboard", allRoles, c.LeaderboardController.GetPublic)
	auth.Get("/groups/:id/leaderboard", allRoles, pc, c.LeaderboardController.GetGroup)

	// ==========================================
	// GAME PROGRESS (member only)
	// ==========================================
	auth.Post("/progress/start", memberRole, pc, c.ProgressController.Start)
	auth.Post("/progress/answer", memberRole, pc, c.ProgressController.Answer)
	auth.Post("/progress/scene", memberRole, pc, c.ProgressController.Scene)
	auth.Post("/progress/solve", memberRole, pc, c.ProgressController.Solve)
	auth.Post("/progress/submit", memberRole, pc, c.ProgressController.Submit)
	auth.Get("/progress/session/:content_type/:content_id", memberRole, pc, c.ProgressController.GetSession)
}
