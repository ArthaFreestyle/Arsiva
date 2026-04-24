package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type CeritaRepository interface {
	FindAll(ctx context.Context, page int, size int, search string) ([]*entity.CeritaInteraktif, int, error)
	FindById(ctx context.Context, ceritaId int) (*entity.CeritaInteraktif, error)
	Create(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error)
	Update(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error)
	CreateScene(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error)
	UpdateScene(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error)
	DeleteScene(ctx context.Context, ceritaId int, sceneId int) error
	Delete(ctx context.Context, ceritaId int) error

	FindAllManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.CeritaInteraktif, int, error)
	FindByIdManage(ctx context.Context, ceritaId int, userId string, role string) (*entity.CeritaInteraktif, error)
}

type ceritaRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewCeritaRepository(db *pgxpool.Pool, log *logrus.Logger) CeritaRepository {
	return &ceritaRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

// FindAll returns paginated cerita interaktif with total count.
func (r *ceritaRepositoryImpl) FindAll(ctx context.Context, page int, size int, search string) ([]*entity.CeritaInteraktif, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM cerita_interaktif WHERE is_published = true AND judul ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT c.cerita_id, c.judul, COALESCE(ass.url,'') AS thumbnail, c.thumbnail_asset_id,
		COALESCE(c.deskripsi,'') AS deskripsi, c.kategori_id, c.xp_reward,
		c.created_at, c.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS "user"
		FROM cerita_interaktif c
		LEFT JOIN users u ON c.created_by = u.user_id
		LEFT JOIN assets ass ON c.thumbnail_asset_id = ass.asset_id
		WHERE c.is_published = true AND c.judul ILIKE $1
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}

	ceritas, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.CeritaInteraktif])
	if err != nil {
		return nil, 0, err
	}
	return ceritas, total, nil
}

// FindById returns a cerita interaktif with all its scenes.
func (r *ceritaRepositoryImpl) FindById(ctx context.Context, ceritaId int) (*entity.CeritaInteraktif, error) {
	// 1. Fetch the cerita
	ceritaSQL := `SELECT c.cerita_id, c.judul, COALESCE(ass.url,'') AS thumbnail, c.thumbnail_asset_id,
		COALESCE(c.deskripsi,'') AS deskripsi, c.kategori_id, c.xp_reward,
		c.created_at, c.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS "user"
		FROM cerita_interaktif c
		LEFT JOIN users u ON c.created_by = u.user_id
		LEFT JOIN assets ass ON c.thumbnail_asset_id = ass.asset_id
		WHERE c.cerita_id = $1`

	rows, err := r.DB.Query(ctx, ceritaSQL, ceritaId)
	if err != nil {
		return nil, err
	}
	cerita, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.CeritaInteraktif])
	if err != nil {
		return nil, err
	}

	// 2. Fetch scenes
	sceneSQL := `SELECT s.scene_id, s.cerita_id, s.scene_key, COALESCE(ass.url,'') AS scene_image, s.scene_image_asset_id,
		s.scene_text, COALESCE(s.scene_choices, '[]'::jsonb) AS scene_choices,
		s.is_ending, s.ending_point, COALESCE(s.ending_type,'') AS ending_type, COALESCE(s.urutan,0) AS urutan
		FROM scene s
		LEFT JOIN assets ass ON s.scene_image_asset_id = ass.asset_id
		WHERE s.cerita_id = $1 ORDER BY s.urutan`

	sRows, err := r.DB.Query(ctx, sceneSQL, ceritaId)
	if err != nil {
		return nil, err
	}
	scenes, err := pgx.CollectRows(sRows, pgx.RowToAddrOfStructByNameLax[entity.Scene])
	if err != nil {
		return nil, err
	}

	cerita.Scenes = scenes
	return cerita, nil
}

// Create inserts a cerita interaktif with its scenes using a transaction.
func (r *ceritaRepositoryImpl) Create(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Insert cerita
	ceritaSQL := `INSERT INTO cerita_interaktif (judul, thumbnail_asset_id, deskripsi, kategori_id, xp_reward, created_by, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING cerita_id, created_at`

	err = tx.QueryRow(ctx, ceritaSQL,
		cerita.Judul, cerita.ThumbnailAssetId, cerita.Deskripsi,
		cerita.KategoriId, cerita.XpReward,
		cerita.CreatedBy.UserId, cerita.IsPublished,
	).Scan(&cerita.CeritaId, &cerita.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 2. Batch insert scenes
	if len(cerita.Scenes) > 0 {
		sceneValues := make([]string, 0, len(cerita.Scenes))
		sceneArgs := make([]interface{}, 0, len(cerita.Scenes)*9)
		argIdx := 1

		for _, s := range cerita.Scenes {
			sceneValues = append(sceneValues,
				fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
					argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5, argIdx+6, argIdx+7, argIdx+8))

			choicesJSON, _ := json.Marshal(s.SceneChoices)

			sceneArgs = append(sceneArgs,
				cerita.CeritaId, s.SceneKey, s.SceneImageAssetId, s.SceneText,
				choicesJSON, s.IsEnding, s.EndingPoint, s.EndingType, s.Urutan)
			argIdx += 9
		}

		sceneSQL := fmt.Sprintf(
			`INSERT INTO scene (cerita_id, scene_key, scene_image_asset_id, scene_text, scene_choices, is_ending, ending_point, ending_type, urutan)
			VALUES %s RETURNING scene_id`, strings.Join(sceneValues, ", "))

		sRows, err := tx.Query(ctx, sceneSQL, sceneArgs...)
		if err != nil {
			return nil, err
		}

		sceneIds, err := pgx.CollectRows(sRows, pgx.RowTo[int])
		if err != nil {
			return nil, err
		}

		for i, s := range cerita.Scenes {
			s.SceneId = sceneIds[i]
			s.CeritaId = cerita.CeritaId
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return r.FindById(ctx, cerita.CeritaId)
}

// Update updates cerita metadata.
func (r *ceritaRepositoryImpl) Update(ctx context.Context, cerita *entity.CeritaInteraktif) (*entity.CeritaInteraktif, error) {
	updateSQL := `UPDATE cerita_interaktif SET judul = $1, thumbnail_asset_id = $2, deskripsi = $3,
		kategori_id = $4, xp_reward = $5, is_published = $6
		WHERE cerita_id = $7`

	_, err := r.DB.Exec(ctx, updateSQL,
		cerita.Judul, cerita.ThumbnailAssetId, cerita.Deskripsi,
		cerita.KategoriId, cerita.XpReward, cerita.IsPublished, cerita.CeritaId)
	if err != nil {
		return nil, err
	}

	return r.FindById(ctx, cerita.CeritaId)
}

func (r *ceritaRepositoryImpl) CreateScene(ctx context.Context, ceritaId int, scene *entity.Scene) (*entity.Scene, error) {
	choicesJSON, _ := json.Marshal(scene.SceneChoices)
	insertSQL := `INSERT INTO scene (cerita_id, scene_key, scene_image_asset_id, scene_text, scene_choices, is_ending, ending_point, ending_type, urutan)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING scene_id`

	err := r.DB.QueryRow(ctx, insertSQL,
		ceritaId, scene.SceneKey, scene.SceneImageAssetId, scene.SceneText, choicesJSON,
		scene.IsEnding, scene.EndingPoint, scene.EndingType, scene.Urutan,
	).Scan(&scene.SceneId)
	if err != nil {
		return nil, err
	}

	return r.findSceneById(ctx, ceritaId, scene.SceneId)
}

func (r *ceritaRepositoryImpl) UpdateScene(ctx context.Context, ceritaId int, sceneId int, scene *entity.Scene) (*entity.Scene, error) {
	choicesJSON, _ := json.Marshal(scene.SceneChoices)
	updateSQL := `UPDATE scene
		SET scene_key = $1, scene_image_asset_id = $2, scene_text = $3, scene_choices = $4,
			is_ending = $5, ending_point = $6, ending_type = $7, urutan = $8
		WHERE scene_id = $9 AND cerita_id = $10`

	_, err := r.DB.Exec(ctx, updateSQL,
		scene.SceneKey, scene.SceneImageAssetId, scene.SceneText, choicesJSON,
		scene.IsEnding, scene.EndingPoint, scene.EndingType, scene.Urutan, sceneId, ceritaId,
	)
	if err != nil {
		return nil, err
	}

	return r.findSceneById(ctx, ceritaId, sceneId)
}

func (r *ceritaRepositoryImpl) DeleteScene(ctx context.Context, ceritaId int, sceneId int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM scene WHERE scene_id = $1 AND cerita_id = $2`, sceneId, ceritaId)
	return err
}

// Delete removes a cerita interaktif (CASCADE handles scenes).
func (r *ceritaRepositoryImpl) Delete(ctx context.Context, ceritaId int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM cerita_interaktif WHERE cerita_id = $1`, ceritaId)
	return err
}

func (r *ceritaRepositoryImpl) findSceneById(ctx context.Context, ceritaId int, sceneId int) (*entity.Scene, error) {
	sceneSQL := `SELECT s.scene_id, s.cerita_id, s.scene_key, COALESCE(ass.url,'') AS scene_image, s.scene_image_asset_id,
		s.scene_text, COALESCE(s.scene_choices, '[]'::jsonb) AS scene_choices,
		s.is_ending, s.ending_point, COALESCE(s.ending_type,'') AS ending_type, COALESCE(s.urutan,0) AS urutan
		FROM scene s
		LEFT JOIN assets ass ON s.scene_image_asset_id = ass.asset_id
		WHERE s.cerita_id = $1 AND s.scene_id = $2`

	rows, err := r.DB.Query(ctx, sceneSQL, ceritaId, sceneId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Scene])
}

func (r *ceritaRepositoryImpl) FindAllManage(ctx context.Context, page int, size int, search string, userId string, role string) ([]*entity.CeritaInteraktif, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	var whereClause string
	var countArgs, queryArgs []interface{}

	if role == "super_admin" {
		whereClause = "WHERE c.judul ILIKE $1"
		countArgs = []interface{}{searchPattern}
		queryArgs = []interface{}{searchPattern, size, offset}
	} else {
		whereClause = "WHERE c.created_by = $1 AND c.judul ILIKE $2"
		countArgs = []interface{}{userId, searchPattern}
		queryArgs = []interface{}{userId, searchPattern, size, offset}
	}

	var total int
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM cerita_interaktif c %s`, whereClause)
	err := r.DB.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limitOffset := fmt.Sprintf("LIMIT $%d OFFSET $%d", len(countArgs)+1, len(countArgs)+2)

	SQL := fmt.Sprintf(`SELECT c.cerita_id, c.judul, COALESCE(ass.url,'') AS thumbnail, c.thumbnail_asset_id,
		COALESCE(c.deskripsi,'') AS deskripsi, c.kategori_id, c.xp_reward,
		c.created_at, c.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS "user"
		FROM cerita_interaktif c
		LEFT JOIN users u ON c.created_by = u.user_id
		LEFT JOIN assets ass ON c.thumbnail_asset_id = ass.asset_id
		%s
		ORDER BY c.created_at DESC
		%s`, whereClause, limitOffset)

	rows, err := r.DB.Query(ctx, SQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	ceritas, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.CeritaInteraktif])
	if err != nil {
		return nil, 0, err
	}
	return ceritas, total, nil
}

func (r *ceritaRepositoryImpl) FindByIdManage(ctx context.Context, ceritaId int, userId string, role string) (*entity.CeritaInteraktif, error) {
	var whereClause string
	var args []interface{}

	if role == "super_admin" {
		whereClause = "WHERE c.cerita_id = $1"
		args = []interface{}{ceritaId}
	} else {
		whereClause = "WHERE c.cerita_id = $1 AND c.created_by = $2"
		args = []interface{}{ceritaId, userId}
	}

	ceritaSQL := fmt.Sprintf(`SELECT c.cerita_id, c.judul, COALESCE(ass.url,'') AS thumbnail, c.thumbnail_asset_id,
		COALESCE(c.deskripsi,'') AS deskripsi, c.kategori_id, c.xp_reward,
		c.created_at, c.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id::text,
			'username', u.username
		) AS "user"
		FROM cerita_interaktif c
		LEFT JOIN users u ON c.created_by = u.user_id
		LEFT JOIN assets ass ON c.thumbnail_asset_id = ass.asset_id
		%s`, whereClause)

	rows, err := r.DB.Query(ctx, ceritaSQL, args...)
	if err != nil {
		return nil, err
	}
	cerita, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.CeritaInteraktif])
	if err != nil {
		return nil, err
	}

	// Fetch scenes
	sceneSQL := `SELECT s.scene_id, s.cerita_id, s.scene_key, COALESCE(ass.url,'') AS scene_image, s.scene_image_asset_id,
		s.scene_text, COALESCE(s.scene_choices, '[]'::jsonb) AS scene_choices,
		s.is_ending, s.ending_point, COALESCE(s.ending_type,'') AS ending_type, COALESCE(s.urutan,0) AS urutan
		FROM scene s
		LEFT JOIN assets ass ON s.scene_image_asset_id = ass.asset_id
		WHERE s.cerita_id = $1 ORDER BY s.urutan`

	sRows, err := r.DB.Query(ctx, sceneSQL, ceritaId)
	if err != nil {
		return nil, err
	}
	scenes, err := pgx.CollectRows(sRows, pgx.RowToAddrOfStructByNameLax[entity.Scene])
	if err != nil {
		return nil, err
	}

	cerita.Scenes = scenes
	return cerita, nil
}
