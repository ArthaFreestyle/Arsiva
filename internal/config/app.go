package config

import (
	"ArthaFreestyle/Arsiva/delivery/http/middleware"
	"ArthaFreestyle/Arsiva/delivery/http/route"
	"ArthaFreestyle/Arsiva/internal/delivery/http"
	"ArthaFreestyle/Arsiva/internal/repository"
	"ArthaFreestyle/Arsiva/internal/usecase"
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type BootstrapConfig struct {
	DB       *pgxpool.Pool
	Redis    *redis.Client
	App      *fiber.App
	Log      *logrus.Logger
	Validate *validator.Validate
	Secret   []byte
	Config   *viper.Viper
}

func Bootstrap(cfg BootstrapConfig) {
	userRepo := repository.NewUserRepository(cfg.DB, cfg.Log)
	articleCategoryRepo := repository.NewArticleCategoryRepository(cfg.DB, cfg.Log)
	articleRepo := repository.NewArticleRepository(cfg.DB, cfg.Log)
	puzzleRepo := repository.NewPuzzleRepository(cfg.DB, cfg.Log)
	quizRepo := repository.NewQuizRepository(cfg.DB, cfg.Log)
	ceritaRepo := repository.NewCeritaRepository(cfg.DB, cfg.Log)
	storyCategoryRepo := repository.NewStoryCategoryRepository(cfg.DB, cfg.Log)
	quizCategoryRepo := repository.NewQuizCategoryRepository(cfg.DB, cfg.Log)
	assetRepo := repository.NewAssetRepository(cfg.DB, cfg.Log)
	groupRepo := repository.NewGroupRepository(cfg.DB, cfg.Log)
	sekolahRepo := repository.NewSekolahRepository(cfg.DB, cfg.Log)
	guruRepo := repository.NewGuruRepository(cfg.DB, cfg.Log)
	memberRepo := repository.NewMemberRepository(cfg.DB, cfg.Log)
	memberAchievementRepo := repository.NewMemberAchievementRepository(cfg.DB, cfg.Log)
	memberSocialLinkRepo := repository.NewMemberSocialLinkRepository(cfg.DB, cfg.Log)
	achievementRepo := repository.NewAchievementRepository(cfg.DB, cfg.Log)
	memberProgressRepo := repository.NewMemberProgressRepository(cfg.DB, cfg.Log)
	leaderboardRepo := repository.NewLeaderboardRepository(cfg.DB, cfg.Log)

	AuthUseCase := usecase.NewAuthUseCase(userRepo, cfg.Secret, cfg.Validate, cfg.Log, cfg.DB, guruRepo, memberRepo)
	UserUseCase := usecase.NewUserUseCase(userRepo, cfg.Log, cfg.DB, cfg.Validate)
	ArticleCategoryUseCase := usecase.NewArticleCategoryUseCase(articleCategoryRepo, cfg.Redis, cfg.Log, cfg.Validate)
	ArticleUseCase := usecase.NewArticleUseCase(articleRepo, assetRepo, cfg.Log, cfg.Validate)
	PuzzleUseCase := usecase.NewPuzzleUseCase(puzzleRepo, assetRepo, cfg.Log, cfg.Validate)
	QuizUseCase := usecase.NewQuizUseCase(quizRepo, assetRepo, cfg.Log, cfg.Validate)
	CeritaUseCase := usecase.NewCeritaUseCase(ceritaRepo, assetRepo, cfg.Log, cfg.Validate)
	StoryCategoryUseCase := usecase.NewStoryCategoryUseCase(storyCategoryRepo, cfg.Log, cfg.Validate)
	QuizCategoryUseCase := usecase.NewQuizCategoryUseCase(quizCategoryRepo, cfg.Log, cfg.Validate)
	AssetUseCase := usecase.NewAssetUsecase(assetRepo, cfg.Log, "./uploads")
	GroupUseCase := usecase.NewGroupUseCase(groupRepo, assetRepo, cfg.Log, cfg.Validate, cfg.Secret)
	SekolahUseCase := usecase.NewSekolahUseCase(sekolahRepo, cfg.Log, cfg.Validate)
	GuruUseCase := usecase.NewGuruUseCase(guruRepo, cfg.Log, cfg.Validate)
	MemberUseCase := usecase.NewMemberUseCase(memberRepo, memberAchievementRepo, memberSocialLinkRepo, cfg.Log, cfg.Validate)
	MemberSocialLinkUseCase := usecase.NewMemberSocialLinkUseCase(memberSocialLinkRepo, cfg.Log, cfg.Validate)
	AchievementUseCase := usecase.NewAchievementUseCase(achievementRepo, cfg.Log, cfg.Validate)
	MemberAchievementUseCase := usecase.NewMemberAchievementUseCase(memberAchievementRepo, memberRepo, achievementRepo, cfg.Log, cfg.Validate)
	ProgressSessionUseCase := usecase.NewProgressSessionUseCase(memberProgressRepo, cfg.Redis, cfg.Log, cfg.Validate)
	LeaderboardUseCase := usecase.NewLeaderboardUseCase(leaderboardRepo, groupRepo, cfg.Log)

	if !fiber.IsChild() {
		cfg.Log.Info("Starting asset cleanup worker on master process...")
		go startAssetCleanupCron(AssetUseCase, cfg.Log)
		cfg.Log.Info("Starting progress flush worker on master process...")
		go startProgressFlushWorker(ProgressSessionUseCase, cfg.Log)
	}

	AuthController := http.NewAuthController(cfg.Log, AuthUseCase)
	UserController := http.NewUserController(UserUseCase, cfg.Log)
	ArticleCategoryController := http.NewArticleCategoryController(ArticleCategoryUseCase, cfg.Log)
	ArticleController := http.NewArticleController(ArticleUseCase, cfg.Log)
	UploadController := http.NewUploadController(cfg.Log, "./uploads", AssetUseCase)
	PuzzleController := http.NewPuzzleController(PuzzleUseCase, cfg.Log)
	QuizController := http.NewQuizController(QuizUseCase, cfg.Log)
	CeritaController := http.NewCeritaController(CeritaUseCase, cfg.Log)
	StoryCategoryController := http.NewStoryCategoryController(StoryCategoryUseCase, cfg.Log)
	QuizCategoryController := http.NewQuizCategoryController(QuizCategoryUseCase, cfg.Log)
	GroupController := http.NewGroupController(GroupUseCase, cfg.Log)
	SekolahController := http.NewSekolahController(SekolahUseCase, cfg.Log)
	GuruController := http.NewGuruController(GuruUseCase, cfg.Log)
	MemberController := http.NewMemberController(MemberUseCase, cfg.Log)
	MemberSocialLinkController := http.NewMemberSocialLinkController(MemberSocialLinkUseCase, cfg.Log)
	AchievementController := http.NewAchievementController(AchievementUseCase, cfg.Log)
	MemberAchievementController := http.NewMemberAchievementController(MemberAchievementUseCase, cfg.Log)
	ProgressController := http.NewProgressController(ProgressSessionUseCase, cfg.Log)
	LeaderboardController := http.NewLeaderboardController(LeaderboardUseCase, cfg.Log)

	// Create middleware
	authMiddleware            := middleware.NewAuthMiddleware(cfg.Secret, cfg.Log)
	profileCompleteMiddleware := middleware.RequireProfileComplete()

	routeConfig := route.RouteConfig{
		App:                       cfg.App,
		AuthController:            AuthController,
		UserController:            UserController,
		ArticleCategoryController: ArticleCategoryController,
		ArticleController:         ArticleController,
		UploadController:          UploadController,
		PuzzleController:          PuzzleController,
		QuizController:            QuizController,
		CeritaController:          CeritaController,
		StoryCategoryController:   StoryCategoryController,
		QuizCategoryController:    QuizCategoryController,
		GroupController:           GroupController,
		SekolahController:         SekolahController,
		GuruController:            GuruController,
		MemberController:            MemberController,
		MemberSocialLinkController:  MemberSocialLinkController,
		MemberAchievementController: MemberAchievementController,
		AchievementController:       AchievementController,
		ProgressController:          ProgressController,
		LeaderboardController:       LeaderboardController,
		AuthMiddleware:              authMiddleware,
		ProfileCompleteMiddleware:   profileCompleteMiddleware,
	}

	routeConfig.SetupRoutes()

}

func startAssetCleanupCron(u usecase.AssetUsecase, log *logrus.Logger) {
	ctx := context.Background()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	go func() {
		log.Info("Executing initial orphaned assets cleanup...")
		if err := u.CleanupOrphanedAssets(ctx); err != nil {
			log.Errorf("Initial asset cleanup failed: %v", err)
		}
	}()

	for range ticker.C {
		log.Info("Running scheduled asset cleanup...")
		if err := u.CleanupOrphanedAssets(ctx); err != nil {
			log.Errorf("Scheduled asset cleanup failed: %v", err)
		}
	}
}
