package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type AssetRepository interface {
	Create(ctx context.Context, url string) (int, error)
	MarkAsUsedTx(ctx context.Context, tx pgx.Tx, assetIds []int) error
	MarkAsUsed(ctx context.Context, assetIds []int) error
	GetOrphanedAssets(ctx context.Context, olderThanDays int) ([]int, []string, error)
	DeleteSoft(ctx context.Context, assetIds []int) error
}

type assetRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewAssetRepository(db *pgxpool.Pool, log *logrus.Logger) AssetRepository {
	return &assetRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *assetRepositoryImpl) Create(ctx context.Context, url string) (int, error) {
	var assetId int
	err := r.DB.QueryRow(ctx, "INSERT INTO assets (url, is_used) VALUES ($1, false) RETURNING asset_id", url).Scan(&assetId)
	return assetId, err
}

func (r *assetRepositoryImpl) MarkAsUsedTx(ctx context.Context, tx pgx.Tx, assetIds []int) error {
	if len(assetIds) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, "UPDATE assets SET is_used = true WHERE asset_id = ANY($1)", assetIds)
	return err
}

func (r *assetRepositoryImpl) MarkAsUsed(ctx context.Context, assetIds []int) error {
	if len(assetIds) == 0 {
		return nil
	}
	_, err := r.DB.Exec(ctx, "UPDATE assets SET is_used = true WHERE asset_id = ANY($1)", assetIds)
	return err
}

func (r *assetRepositoryImpl) GetOrphanedAssets(ctx context.Context, olderThanDays int) ([]int, []string, error) {
	query := `SELECT asset_id, url FROM assets WHERE is_used = false AND created_at < NOW() - INTERVAL '1 day' * $1 AND deleted_at IS NULL`
	rows, err := r.DB.Query(ctx, query, olderThanDays)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids []int
	var urls []string
	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		urls = append(urls, url)
	}
	return ids, urls, nil
}

func (r *assetRepositoryImpl) DeleteSoft(ctx context.Context, assetIds []int) error {
	if len(assetIds) == 0 {
		return nil
	}
	_, err := r.DB.Exec(ctx, "UPDATE assets SET deleted_at = NOW() WHERE asset_id = ANY($1)", assetIds)
	return err
}
