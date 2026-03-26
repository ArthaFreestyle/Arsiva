package config

import (
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
	"ArthaFreestyle/Arsiva/delivery/http/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/gofiber/fiber/v3"
	"ArthaFreestyle/Arsiva/delivery/http/route"
)

type BootstrapConfig struct{
	DB				*pgxpool.Pool
	App				*fiber.App
	Log				*logrus.Logger
	Validate		*validator.Validate
	Secret			[]byte
	Config			*viper.Viper
}

func Bootstrap(cfg BootstrapConfig) {
	userRepo := repository.NewUserRepository(cfg.DB,cfg.Log)
	articleCategoryRepo := repository.NewArticleCategoryRepository(cfg.DB,cfg.Log)
	articleRepo := repository.NewArticleRepository(cfg.DB,cfg.Log)
	puzzleRepo := repository.NewPuzzleRepository(cfg.DB,cfg.Log)
	quizRepo := repository.NewQuizRepository(cfg.DB,cfg.Log)
	ceritaRepo := repository.NewCeritaRepository(cfg.DB,cfg.Log)
	storyCategoryRepo := repository.NewStoryCategoryRepository(cfg.DB,cfg.Log)

	AuthUseCase := usecase.NewAuthUseCase(userRepo,cfg.Secret,cfg.Validate,cfg.Log,cfg.DB)
	UserUseCase := usecase.NewUserUseCase(userRepo,cfg.Log,cfg.DB,cfg.Validate)
	ArticleCategoryUseCase := usecase.NewArticleCategoryUseCase(articleCategoryRepo,cfg.Log,cfg.Validate)
	ArticleUseCase := usecase.NewArticleUseCase(articleRepo,cfg.Log,cfg.Validate)
	PuzzleUseCase := usecase.NewPuzzleUseCase(puzzleRepo,cfg.Log,cfg.Validate)
	QuizUseCase := usecase.NewQuizUseCase(quizRepo,cfg.Log,cfg.Validate)
	CeritaUseCase := usecase.NewCeritaUseCase(ceritaRepo,cfg.Log,cfg.Validate)
	StoryCategoryUseCase := usecase.NewStoryCategoryUseCase(storyCategoryRepo,cfg.Log,cfg.Validate)

	AuthController := http.NewAuthController(cfg.Log,AuthUseCase)
	UserController := http.NewUserController(UserUseCase,cfg.Log)
	ArticleCategoryController := http.NewArticleCategoryController(ArticleCategoryUseCase,cfg.Log)
	ArticleController := http.NewArticleController(ArticleUseCase,cfg.Log)
	UploadController := http.NewUploadController(cfg.Log,"./uploads")
	PuzzleController := http.NewPuzzleController(PuzzleUseCase,cfg.Log)
	QuizController := http.NewQuizController(QuizUseCase,cfg.Log)
	CeritaController := http.NewCeritaController(CeritaUseCase,cfg.Log)
	StoryCategoryController := http.NewStoryCategoryController(StoryCategoryUseCase,cfg.Log)

	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Secret,cfg.Log)

	routeConfig := route.RouteConfig{
		App: cfg.App,
		AuthController : AuthController,
		UserController : UserController,
		ArticleCategoryController : ArticleCategoryController,
		ArticleController : ArticleController,
		UploadController : UploadController,
		PuzzleController : PuzzleController,
		QuizController : QuizController,
		CeritaController : CeritaController,
		StoryCategoryController : StoryCategoryController,
		AuthMiddleware : authMiddleware,
	}

	routeConfig.SetupRoutes()
	
}