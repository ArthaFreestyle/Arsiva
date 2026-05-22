package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ErrGroupContentNotFound is returned when a group_content row doesn't exist or doesn't belong to the given group.
var ErrGroupContentNotFound = errors.New("group content not found")

type GroupRepository interface {
	// Group CRUD
	CreateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error)
	GetAllGroupsByGuru(ctx context.Context, guruId int, page int, size int, search string) ([]*entity.Group, int, error)
	GetAllGroupsByMember(ctx context.Context, memberId int, page int, size int, search string) ([]*entity.Group, int, error)
	GetGroupById(ctx context.Context, groupId string) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupId string) error

	// Group Members
	AddMember(ctx context.Context, groupId string, memberId int) error
	RemoveMember(ctx context.Context, groupId string, memberId int) error
	GetGroupMembers(ctx context.Context, groupId string) ([]*entity.GroupMember, error)
	IsMemberInGroup(ctx context.Context, groupId string, memberId int) (bool, error)

	// Group Contents
	AddContent(ctx context.Context, groupId string, contentType string, contentId int) (*entity.GroupContent, error)
	GetContentsByGroupId(ctx context.Context, groupId string, contentType string) ([]*entity.GroupContent, error)
	RemoveContent(ctx context.Context, groupContentId int, groupId string) error
	ContentExists(ctx context.Context, contentType string, contentId int) (bool, error)
	IsContentAlreadyAssigned(ctx context.Context, groupId string, contentType string, contentId int) (bool, error)

	// Helper
	GetGuruIdByUserId(ctx context.Context, userId string) (int, error)
	GetMemberIdByEmail(ctx context.Context, email string) (int, error)
	GetMemberIdByUserId(ctx context.Context, userId string) (int, error)
}

type groupRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewGroupRepository(db *pgxpool.Pool, log *logrus.Logger) GroupRepository {
	return &groupRepositoryImpl{
		DB:  db,
		Log: log,
	}
}

func (r *groupRepositoryImpl) CreateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error) {
	group.GroupId = uuid.New().String()
	query := `INSERT INTO groups (group_id, group_name, created_by, created_at, updated_at) 
	          VALUES ($1, $2, $3, NOW(), NOW()) RETURNING group_id`

	r.Log.Infof("Executing CreateGroup query")
	var id string
	err := r.DB.QueryRow(ctx, query, group.GroupId, group.GroupName, group.CreatedBy).Scan(&id)
	if err != nil {
		r.Log.Errorf("Error CreateGroup: %v", err)
		return nil, err
	}

	return r.GetGroupById(ctx, group.GroupId)
}

func (r *groupRepositoryImpl) GetAllGroupsByGuru(ctx context.Context, guruId int, page int, size int, search string) ([]*entity.Group, int, error) {
	searchPattern := "%" + search + "%"
	limit := size
	offset := (page - 1) * size

	countQuery := `SELECT COUNT(*) FROM groups WHERE created_by = $1 AND group_name ILIKE $2`
	var total int
	err := r.DB.QueryRow(ctx, countQuery, guruId, searchPattern).Scan(&total)
	if err != nil {
		r.Log.Errorf("Error counting GetAllGroupsByGuru: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT 
			g.group_id,
			g.group_name,
			COALESCE(a.url, '') AS group_thumbnail,
			g.group_thumbnail_asset_id,
			g.created_by,
			g.created_at,
			g.updated_at,
			JSON_BUILD_OBJECT(
				'GuruId', gu.guru_id::text,
				'NIP', gu.nip,
				'BidangAjar', gu.bidang_ajar,
				'Username', u.username
			) AS guru,
			(SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.group_id) AS member_count
		FROM groups g
		LEFT JOIN guru gu ON g.created_by = gu.guru_id
		LEFT JOIN users u ON gu.user_id = u.user_id
		LEFT JOIN assets a ON g.group_thumbnail_asset_id = a.asset_id
		WHERE g.created_by = $1 AND g.group_name ILIKE $2
		ORDER BY g.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.DB.Query(ctx, query, guruId, searchPattern, limit, offset)
	if err != nil {
		r.Log.Errorf("Error query GetAllGroupsByGuru: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	groups, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Group])
	if err != nil {
		r.Log.Errorf("Error collecting rows GetAllGroupsByGuru: %v", err)
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *groupRepositoryImpl) GetAllGroupsByMember(ctx context.Context, memberId int, page int, size int, search string) ([]*entity.Group, int, error) {
	searchPattern := "%" + search + "%"
	limit := size
	offset := (page - 1) * size

	countQuery := `SELECT COUNT(*) FROM groups g JOIN group_members gm ON gm.group_id = g.group_id WHERE gm.member_id = $1 AND g.group_name ILIKE $2`
	var total int
	err := r.DB.QueryRow(ctx, countQuery, memberId, searchPattern).Scan(&total)
	if err != nil {
		r.Log.Errorf("Error counting GetAllGroupsByMember: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT
			g.group_id,
			g.group_name,
			COALESCE(a.url, '') AS group_thumbnail,
			g.group_thumbnail_asset_id,
			g.created_by,
			g.created_at,
			g.updated_at,
			JSON_BUILD_OBJECT(
				'GuruId', gu.guru_id::text,
				'NIP', gu.nip,
				'BidangAjar', gu.bidang_ajar,
				'Username', u.username
			) AS guru,
			(SELECT COUNT(*) FROM group_members gm2 WHERE gm2.group_id = g.group_id) AS member_count
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.group_id
		LEFT JOIN guru gu ON g.created_by = gu.guru_id
		LEFT JOIN users u ON gu.user_id = u.user_id
		LEFT JOIN assets a ON g.group_thumbnail_asset_id = a.asset_id
		WHERE gm.member_id = $1 AND g.group_name ILIKE $2
		ORDER BY g.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.DB.Query(ctx, query, memberId, searchPattern, limit, offset)
	if err != nil {
		r.Log.Errorf("Error query GetAllGroupsByMember: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	groups, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.Group])
	if err != nil {
		r.Log.Errorf("Error collecting rows GetAllGroupsByMember: %v", err)
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *groupRepositoryImpl) GetGroupById(ctx context.Context, groupId string) (*entity.Group, error) {
	query := `
		SELECT 
			g.group_id,
			g.group_name,
			COALESCE(a.url, '') AS group_thumbnail,
			g.group_thumbnail_asset_id,
			g.created_by,
			g.created_at,
			g.updated_at,
			JSON_BUILD_OBJECT(
				'GuruId', gu.guru_id::text,
				'NIP', gu.nip,
				'BidangAjar', gu.bidang_ajar,
				'Username', u.username
			) AS guru,
			(SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.group_id) AS member_count
		FROM groups g
		LEFT JOIN guru gu ON g.created_by = gu.guru_id
		LEFT JOIN users u ON gu.user_id = u.user_id
		LEFT JOIN assets a ON g.group_thumbnail_asset_id = a.asset_id
		WHERE g.group_id = $1
	`

	rows, err := r.DB.Query(ctx, query, groupId)
	if err != nil {
		r.Log.Errorf("Error query GetGroupById: %v", err)
		return nil, err
	}
	defer rows.Close()

	group, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.Group])
	if err != nil {
		r.Log.Errorf("Error collecting row GetGroupById: %v", err)
		return nil, err
	}

	return group, nil
}

func (r *groupRepositoryImpl) UpdateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error) {
	query := `UPDATE groups SET group_name = $1, group_thumbnail_asset_id = $2, updated_at = NOW() 
	          WHERE group_id = $3 RETURNING group_id`

	var id string
	err := r.DB.QueryRow(ctx, query, group.GroupName, group.GroupThumbnailAssetId, group.GroupId).Scan(&id)
	if err != nil {
		r.Log.Errorf("Error UpdateGroup: %v", err)
		return nil, err
	}

	return r.GetGroupById(ctx, group.GroupId)
}

func (r *groupRepositoryImpl) DeleteGroup(ctx context.Context, groupId string) error {
	query := `DELETE FROM groups WHERE group_id = $1`
	_, err := r.DB.Exec(ctx, query, groupId)
	if err != nil {
		r.Log.Errorf("Error DeleteGroup: %v", err)
		return err
	}
	return nil
}

func (r *groupRepositoryImpl) AddMember(ctx context.Context, groupId string, memberId int) error {
	query := `INSERT INTO group_members (group_id, member_id, tanggal_bergabung) 
	          VALUES ($1, $2, NOW()) ON CONFLICT (group_id, member_id) DO NOTHING`
	_, err := r.DB.Exec(ctx, query, groupId, memberId)
	if err != nil {
		r.Log.Errorf("Error AddMember: %v", err)
		return err
	}
	return nil
}

func (r *groupRepositoryImpl) RemoveMember(ctx context.Context, groupId string, memberId int) error {
	query := `DELETE FROM group_members WHERE group_id = $1 AND member_id = $2`
	_, err := r.DB.Exec(ctx, query, groupId, memberId)
	if err != nil {
		r.Log.Errorf("Error RemoveMember: %v", err)
		return err
	}
	return nil
}

func (r *groupRepositoryImpl) GetGroupMembers(ctx context.Context, groupId string) ([]*entity.GroupMember, error) {
	query := `
		SELECT 
			gm.group_id,
			gm.member_id,
			gm.tanggal_bergabung,
			u.username,
			u.email,
			m.nis,
			m.foto_profil
		FROM group_members gm
		JOIN members m ON gm.member_id = m.member_id
		JOIN users u ON m.user_id = u.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.tanggal_bergabung ASC
	`

	rows, err := r.DB.Query(ctx, query, groupId)
	if err != nil {
		r.Log.Errorf("Error query GetGroupMembers: %v", err)
		return nil, err
	}
	defer rows.Close()

	members, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.GroupMember])
	if err != nil {
		r.Log.Errorf("Error collecting rows GetGroupMembers: %v", err)
		return nil, err
	}

	return members, nil
}

func (r *groupRepositoryImpl) IsMemberInGroup(ctx context.Context, groupId string, memberId int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND member_id = $2)`
	var exists bool
	err := r.DB.QueryRow(ctx, query, groupId, memberId).Scan(&exists)
	if err != nil {
		r.Log.Errorf("Error IsMemberInGroup: %v", err)
		return false, err
	}
	return exists, nil
}

func (r *groupRepositoryImpl) GetGuruIdByUserId(ctx context.Context, userId string) (int, error) {
	query := `SELECT guru_id FROM guru WHERE user_id = $1`
	var guruId int
	err := r.DB.QueryRow(ctx, query, userId).Scan(&guruId)
	if err != nil {
		r.Log.Errorf("Error GetGuruIdByUserId: %v", err)
		return 0, err
	}
	return guruId, nil
}

func (r *groupRepositoryImpl) GetMemberIdByEmail(ctx context.Context, email string) (int, error) {
	query := `SELECT m.member_id FROM members m JOIN users u ON m.user_id = u.user_id WHERE u.email = $1`
	var memberId int
	err := r.DB.QueryRow(ctx, query, email).Scan(&memberId)
	if err != nil {
		r.Log.Errorf("Error GetMemberIdByEmail: %v", err)
		return 0, err
	}
	return memberId, nil
}

func (r *groupRepositoryImpl) GetMemberIdByUserId(ctx context.Context, userId string) (int, error) {
	query := `SELECT member_id FROM members WHERE user_id = $1`
	var memberId int
	err := r.DB.QueryRow(ctx, query, userId).Scan(&memberId)
	if err != nil {
		r.Log.Errorf("Error GetMemberIdByUserId: %v", err)
		return 0, err
	}
	return memberId, nil
}

// contentJoinQuery is the base SELECT + FROM + LEFT JOINs shared by AddContent and GetContentsByGroupId.
// It joins group_contents with kuis, cerita_interaktif, and puzzles to resolve judul and thumbnail.
const contentJoinQuery = `
	SELECT
		gc.group_content_id,
		gc.group_id,
		gc.content_type,
		gc.content_id,
		CASE
			WHEN gc.content_type = 'kuis'   THEN k.judul
			WHEN gc.content_type = 'cerita' THEN c.judul
			WHEN gc.content_type = 'puzzle' THEN p.judul
		END AS judul,
		COALESCE(
			CASE
				WHEN gc.content_type = 'kuis'   THEN ka.url
				WHEN gc.content_type = 'cerita' THEN ca.url
				WHEN gc.content_type = 'puzzle' THEN pa.url
			END, ''
		) AS thumbnail
	FROM group_contents gc
	LEFT JOIN kuis k              ON gc.content_type = 'kuis'   AND gc.content_id = k.kuis_id
	LEFT JOIN assets ka           ON k.thumbnail_asset_id = ka.asset_id
	LEFT JOIN cerita_interaktif c ON gc.content_type = 'cerita' AND gc.content_id = c.cerita_id
	LEFT JOIN assets ca           ON c.thumbnail_asset_id = ca.asset_id
	LEFT JOIN puzzles p           ON gc.content_type = 'puzzle' AND gc.content_id = p.puzzle_id
	LEFT JOIN assets pa           ON p.thumbnail_asset_id = pa.asset_id
`

func (r *groupRepositoryImpl) AddContent(ctx context.Context, groupId string, contentType string, contentId int) (*entity.GroupContent, error) {
	insertQuery := `INSERT INTO group_contents (group_id, content_type, content_id) VALUES ($1, $2, $3) RETURNING group_content_id`

	r.Log.Infof("Executing AddContent query")
	var groupContentId int
	err := r.DB.QueryRow(ctx, insertQuery, groupId, contentType, contentId).Scan(&groupContentId)
	if err != nil {
		r.Log.Errorf("Error AddContent insert: %v", err)
		return nil, err
	}

	return r.getContentByGroupContentId(ctx, groupContentId)
}

func (r *groupRepositoryImpl) getContentByGroupContentId(ctx context.Context, groupContentId int) (*entity.GroupContent, error) {
	query := contentJoinQuery + `WHERE gc.group_content_id = $1`

	rows, err := r.DB.Query(ctx, query, groupContentId)
	if err != nil {
		r.Log.Errorf("Error query getContentByGroupContentId: %v", err)
		return nil, err
	}
	defer rows.Close()

	content, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.GroupContent])
	if err != nil {
		r.Log.Errorf("Error collecting row getContentByGroupContentId: %v", err)
		return nil, err
	}

	return content, nil
}

func (r *groupRepositoryImpl) GetContentsByGroupId(ctx context.Context, groupId string, contentType string) ([]*entity.GroupContent, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if contentType != "" {
		query := contentJoinQuery + `WHERE gc.group_id = $1 AND gc.content_type = $2`
		rows, err = r.DB.Query(ctx, query, groupId, contentType)
	} else {
		query := contentJoinQuery + `WHERE gc.group_id = $1`
		rows, err = r.DB.Query(ctx, query, groupId)
	}

	if err != nil {
		r.Log.Errorf("Error query GetContentsByGroupId: %v", err)
		return nil, err
	}
	defer rows.Close()

	contents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.GroupContent])
	if err != nil {
		r.Log.Errorf("Error collecting rows GetContentsByGroupId: %v", err)
		return nil, err
	}

	return contents, nil
}

func (r *groupRepositoryImpl) RemoveContent(ctx context.Context, groupContentId int, groupId string) error {
	query := `DELETE FROM group_contents WHERE group_content_id = $1 AND group_id = $2`
	cmdTag, err := r.DB.Exec(ctx, query, groupContentId, groupId)
	if err != nil {
		r.Log.Errorf("Error RemoveContent: %v", err)
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrGroupContentNotFound
	}
	return nil
}

func (r *groupRepositoryImpl) ContentExists(ctx context.Context, contentType string, contentId int) (bool, error) {
	var query string
	switch contentType {
	case "kuis":
		query = `SELECT EXISTS(SELECT 1 FROM kuis WHERE kuis_id = $1 AND is_published = true)`
	case "cerita":
		query = `SELECT EXISTS(SELECT 1 FROM cerita_interaktif WHERE cerita_id = $1 AND is_published = true)`
	case "puzzle":
		query = `SELECT EXISTS(SELECT 1 FROM puzzles WHERE puzzle_id = $1 AND is_published = true)`
	}

	var exists bool
	err := r.DB.QueryRow(ctx, query, contentId).Scan(&exists)
	if err != nil {
		r.Log.Errorf("Error ContentExists: %v", err)
		return false, err
	}
	return exists, nil
}

func (r *groupRepositoryImpl) IsContentAlreadyAssigned(ctx context.Context, groupId string, contentType string, contentId int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_contents WHERE group_id = $1 AND content_type = $2 AND content_id = $3)`
	var exists bool
	err := r.DB.QueryRow(ctx, query, groupId, contentType, contentId).Scan(&exists)
	if err != nil {
		r.Log.Errorf("Error IsContentAlreadyAssigned: %v", err)
		return false, err
	}
	return exists, nil
}
