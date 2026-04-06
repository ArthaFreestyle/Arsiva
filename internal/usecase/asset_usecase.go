package usecase

import (
	"ArthaFreestyle/Arsiva/internal/model"
	"ArthaFreestyle/Arsiva/internal/repository"
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type AssetUsecase interface {
	CreateAsset(ctx context.Context, url string) (*model.AssetResponse, error)
	CleanupOrphanedAssets(ctx context.Context) error
}

type assetUsecaseImpl struct {
	Repo      repository.AssetRepository
	Log       *logrus.Logger
	UploadDir string
}

func NewAssetUsecase(repo repository.AssetRepository, log *logrus.Logger, uploadDir string) AssetUsecase {
	return &assetUsecaseImpl{
		Repo:      repo,
		Log:       log,
		UploadDir: uploadDir,
	}
}

func (u *assetUsecaseImpl) CreateAsset(ctx context.Context, url string) (*model.AssetResponse, error) {
	assetId, err := u.Repo.Create(ctx, url)
	if err != nil {
		u.Log.Warnf("Failed to create asset in db: %+v", err)
		return nil, err
	}
	return &model.AssetResponse{
		AssetId: assetId,
		Url:     url,
	}, nil
}

func (u *assetUsecaseImpl) CleanupOrphanedAssets(ctx context.Context) error {
	u.Log.Info("Starting orphaned assets cleanup...")
	
	ids, urls, err := u.Repo.GetOrphanedAssets(ctx, 7) // Clean assets older than 7 days
	if err != nil {
		u.Log.Errorf("Failed to get orphaned assets: %v", err)
		return err
	}

	if len(ids) == 0 {
		u.Log.Info("No orphaned assets found.")
		return nil
	}

	u.Log.Infof("Found %d orphaned assets. Deleting files...", len(ids))

	// Delete files
	for _, url := range urls {
		// URL is like /uploads/2026/04/uuid.webp
		// We need to map it to UploadDir
		// If UploadDir is ./uploads, then /uploads/ part should be removed or handled.
		// Looking at UploadController: url := fmt.Sprintf("/uploads/%s/%s", subDir, filename)
		// And GetFile: ctx.SendFile("./uploads/"+ctx.Params("*"))
		
		relPath := strings.TrimPrefix(url, "/uploads")
		filePath := filepath.Join(u.UploadDir, relPath)
		
		if err := os.Remove(filePath); err != nil {
			if !os.IsNotExist(err) {
				u.Log.Warnf("Failed to delete file %s: %v", filePath, err)
			}
		} else {
			u.Log.Infof("Deleted file: %s", filePath)
		}
	}

	// Delete from DB
	if err := u.Repo.DeleteSoft(ctx, ids); err != nil {
		u.Log.Errorf("Failed to mark assets as deleted in DB: %v", err)
		return err
	}

	u.Log.Infof("Successfully cleaned up %d orphaned assets.", len(ids))
	return nil
}
