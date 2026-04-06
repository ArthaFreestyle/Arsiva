package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type QuizRepository interface {
	GetAll(ctx context.Context, page int, size int, search string) ([]*entity.Quiz, int, error)
	GetByID(ctx context.Context, quizId int) (*entity.Quiz, error)
	Create(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error)
	Update(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error)
	Delete(ctx context.Context, quizId int) error
}

type quizRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewQuizRepository(db *pgxpool.Pool, log *logrus.Logger) QuizRepository {
	return &quizRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

// GetAll returns paginated quizzes with total count.
func (r *quizRepositoryImpl) GetAll(ctx context.Context, page int, size int, search string) ([]*entity.Quiz, int, error) {
	offset := (page - 1) * size
	searchPattern := "%" + search + "%"

	// Count total
	var total int
	err := r.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM kuis WHERE is_published = true AND judul ILIKE $1`,
		searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	SQL := `SELECT k.kuis_id, k.judul, COALESCE(ass_t.url, '') AS thumbnail, k.thumbnail_asset_id, COALESCE(ass_g.url, '') AS gambar, k.gambar_asset_id, k.xp_reward,
		k.kategori_id,
		COALESCE(kk.nama_kategori, '') AS kategori,
		k.created_at, k.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id,
			'username', u.username
		) AS "user"
		FROM kuis k
		LEFT JOIN users u ON k.created_by = u.user_id
		LEFT JOIN kategori_kuis kk ON k.kategori_id = kk.kategori_id
		LEFT JOIN assets ass_t ON k.thumbnail_asset_id = ass_t.asset_id
		LEFT JOIN assets ass_g ON k.gambar_asset_id = ass_g.asset_id
		WHERE k.is_published = true AND k.judul ILIKE $1
		ORDER BY k.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, SQL, searchPattern, size, offset)
	if err != nil {
		return nil, 0, err
	}

	quizzes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	if err != nil {
		return nil, 0, err
	}
	return quizzes, total, nil
}

// GetByID returns a quiz with its questions and options.
func (r *quizRepositoryImpl) GetByID(ctx context.Context, quizId int) (*entity.Quiz, error) {
	// 1. Fetch the quiz
	quizSQL := `SELECT k.kuis_id, k.judul, COALESCE(ass_g.url, '') AS gambar, k.gambar_asset_id, COALESCE(ass_t.url, '') AS thumbnail, k.thumbnail_asset_id, k.xp_reward,
		k.kategori_id,
		COALESCE(kk.nama_kategori, '') AS kategori,
		k.created_at, k.is_published,
		JSON_BUILD_OBJECT(
			'user_id', u.user_id,
			'username', u.username
		) AS "user"
		FROM kuis k
		LEFT JOIN users u ON k.created_by = u.user_id
		LEFT JOIN kategori_kuis kk ON k.kategori_id = kk.kategori_id
		LEFT JOIN assets ass_t ON k.thumbnail_asset_id = ass_t.asset_id
		LEFT JOIN assets ass_g ON k.gambar_asset_id = ass_g.asset_id
		WHERE k.kuis_id = $1`

	rows, err := r.DB.Query(ctx, quizSQL, quizId)
	if err != nil {
		return nil, err
	}
	quiz, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Quiz])
	if err != nil {
		return nil, err
	}

	// 2. Fetch questions
	questionSQL := `SELECT p.pertanyaan_id, p.kuis_id, p.teks_pertanyaan, COALESCE(ass.url,'') AS image, p.image_asset_id, p.tipe, p.poin, p.urutan
		FROM pertanyaan_kuis p
		LEFT JOIN assets ass ON p.image_asset_id = ass.asset_id
		WHERE p.kuis_id = $1 ORDER BY p.urutan`

	qRows, err := r.DB.Query(ctx, questionSQL, quizId)
	if err != nil {
		return nil, err
	}
	questions, err := pgx.CollectRows(qRows, pgx.RowToAddrOfStructByNameLax[entity.Question])
	if err != nil {
		return nil, err
	}

	if len(questions) > 0 {
		// 3. Collect all pertanyaan_ids and fetch options in a single query
		pertanyaanIds := make([]int, len(questions))
		questionMap := make(map[int]*entity.Question, len(questions))
		for i, q := range questions {
			pertanyaanIds[i] = q.PertanyaanId
			questionMap[q.PertanyaanId] = q
		}

		optionSQL := `SELECT jawaban_id, pertanyaan_id, teks_jawaban, score
			FROM pilihan_kuis WHERE pertanyaan_id = ANY($1) ORDER BY jawaban_id`

		oRows, err := r.DB.Query(ctx, optionSQL, pertanyaanIds)
		if err != nil {
			return nil, err
		}
		options, err := pgx.CollectRows(oRows, pgx.RowToAddrOfStructByNameLax[entity.Option])
		if err != nil {
			return nil, err
		}

		// Assign options to the correct question
		for _, o := range options {
			if q, ok := questionMap[o.PertanyaanId]; ok {
				q.Pilihan = append(q.Pilihan, o)
			}
		}
	}

	quiz.Soal = questions
	return quiz, nil
}

// Create inserts a quiz with its questions and options using a transaction + batch.
func (r *quizRepositoryImpl) Create(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Insert quiz
	quizSQL := `INSERT INTO kuis (judul, gambar_asset_id, thumbnail_asset_id, kategori_id, xp_reward, created_by, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING kuis_id, created_at`

	err = tx.QueryRow(ctx, quizSQL,
		quiz.Judul, quiz.GambarAssetId, quiz.ThumbnailAssetId,
		quiz.KategoriId, quiz.XpReward,
		quiz.CreatedBy.UserId, quiz.IsPublished,
	).Scan(&quiz.QuizId, &quiz.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 2. Batch insert questions
	if len(quiz.Soal) > 0 {
		questionValues := make([]string, 0, len(quiz.Soal))
		questionArgs := make([]interface{}, 0, len(quiz.Soal)*6)
		argIdx := 1

		for _, q := range quiz.Soal {
			questionValues = append(questionValues,
				fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5))
			questionArgs = append(questionArgs, quiz.QuizId, q.TeksPertanyaan, q.ImageAssetId, q.Tipe, q.Poin, q.Urutan)
			argIdx += 6
		}

		questionSQL := fmt.Sprintf(
			`INSERT INTO pertanyaan_kuis (kuis_id, teks_pertanyaan, image_asset_id, tipe, poin, urutan)
			VALUES %s RETURNING pertanyaan_id`, strings.Join(questionValues, ", "))

		qRows, err := tx.Query(ctx, questionSQL, questionArgs...)
		if err != nil {
			return nil, err
		}

		pertanyaanIds, err := pgx.CollectRows(qRows, pgx.RowTo[int])
		if err != nil {
			return nil, err
		}

		// Assign returned IDs back and collect all options for batch insert
		optionValues := make([]string, 0)
		optionArgs := make([]interface{}, 0)
		optArgIdx := 1

		for i, q := range quiz.Soal {
			q.PertanyaanId = pertanyaanIds[i]
			q.KuisId = quiz.QuizId

			for _, o := range q.Pilihan {
				optionValues = append(optionValues,
					fmt.Sprintf("($%d, $%d, $%d)", optArgIdx, optArgIdx+1, optArgIdx+2))
				optionArgs = append(optionArgs, q.PertanyaanId, o.TeksJawaban, o.Score)
				optArgIdx += 3
			}
		}

		// 3. Batch insert all options at once
		if len(optionValues) > 0 {
			optionSQL := fmt.Sprintf(
				`INSERT INTO pilihan_kuis (pertanyaan_id, teks_jawaban, score)
				VALUES %s RETURNING jawaban_id, pertanyaan_id`, strings.Join(optionValues, ", "))

			oRows, err := tx.Query(ctx, optionSQL, optionArgs...)
			if err != nil {
				return nil, err
			}

			type optionResult struct {
				JawabanId    int `db:"jawaban_id"`
				PertanyaanId int `db:"pertanyaan_id"`
			}
			results, err := pgx.CollectRows(oRows, pgx.RowToStructByPos[optionResult])
			if err != nil {
				return nil, err
			}

			// Assign jawaban_id back to options
			questionMap := make(map[int]*entity.Question, len(quiz.Soal))
			for _, q := range quiz.Soal {
				questionMap[q.PertanyaanId] = q
			}
			// reset pilihan to rebuild with IDs
			for _, q := range quiz.Soal {
				q.Pilihan = make([]*entity.Option, 0)
			}

			optIdx := 0
			for _, res := range results {
				if q, ok := questionMap[res.PertanyaanId]; ok {
					// Find the original option data
					origQ := quiz.Soal[0]
					for _, sq := range quiz.Soal {
						if sq.PertanyaanId == res.PertanyaanId {
							origQ = sq
							break
						}
					}
					_ = origQ
					q.Pilihan = append(q.Pilihan, &entity.Option{
						JawabanId:    res.JawabanId,
						PertanyaanId: res.PertanyaanId,
					})
					optIdx++
				}
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return quiz, nil
}

// Update updates quiz metadata and replaces all questions+options.
func (r *quizRepositoryImpl) Update(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Update quiz metadata
	updateSQL := `UPDATE kuis SET judul = $1, gambar_asset_id = $2, thumbnail_asset_id = $3, kategori_id = $4,
		xp_reward = $5, is_published = $6 WHERE kuis_id = $7`

	_, err = tx.Exec(ctx, updateSQL,
		quiz.Judul, quiz.GambarAssetId, quiz.ThumbnailAssetId,
		quiz.KategoriId, quiz.XpReward, quiz.IsPublished, quiz.QuizId)
	if err != nil {
		return nil, err
	}

	// 2. Delete existing questions (CASCADE deletes pilihan_kuis too)
	_, err = tx.Exec(ctx, `DELETE FROM pertanyaan_kuis WHERE kuis_id = $1`, quiz.QuizId)
	if err != nil {
		return nil, err
	}

	// 3. Re-insert questions + options (same batch logic as Create)
	if len(quiz.Soal) > 0 {
		questionValues := make([]string, 0, len(quiz.Soal))
		questionArgs := make([]interface{}, 0, len(quiz.Soal)*6)
		argIdx := 1

		for _, q := range quiz.Soal {
			questionValues = append(questionValues,
				fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5))
			questionArgs = append(questionArgs, quiz.QuizId, q.TeksPertanyaan, q.ImageAssetId, q.Tipe, q.Poin, q.Urutan)
			argIdx += 6
		}

		questionSQL := fmt.Sprintf(
			`INSERT INTO pertanyaan_kuis (kuis_id, teks_pertanyaan, image_asset_id, tipe, poin, urutan)
			VALUES %s RETURNING pertanyaan_id`, strings.Join(questionValues, ", "))

		qRows, err := tx.Query(ctx, questionSQL, questionArgs...)
		if err != nil {
			return nil, err
		}

		pertanyaanIds, err := pgx.CollectRows(qRows, pgx.RowTo[int])
		if err != nil {
			return nil, err
		}

		optionValues := make([]string, 0)
		optionArgs := make([]interface{}, 0)
		optArgIdx := 1

		for i, q := range quiz.Soal {
			q.PertanyaanId = pertanyaanIds[i]
			q.KuisId = quiz.QuizId

			for _, o := range q.Pilihan {
				optionValues = append(optionValues,
					fmt.Sprintf("($%d, $%d, $%d)", optArgIdx, optArgIdx+1, optArgIdx+2))
				optionArgs = append(optionArgs, q.PertanyaanId, o.TeksJawaban, o.Score)
				optArgIdx += 3
			}
		}

		if len(optionValues) > 0 {
			optionSQL := fmt.Sprintf(
				`INSERT INTO pilihan_kuis (pertanyaan_id, teks_jawaban, score)
				VALUES %s`, strings.Join(optionValues, ", "))

			_, err = tx.Exec(ctx, optionSQL, optionArgs...)
			if err != nil {
				return nil, err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return quiz, nil
}

// Delete removes a quiz (CASCADE deletes questions and options).
func (r *quizRepositoryImpl) Delete(ctx context.Context, quizId int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM kuis WHERE kuis_id = $1`, quizId)
	return err
}