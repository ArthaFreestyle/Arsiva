package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type GroupRepository interface {
	// Group CRUD
	CreateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error)
	GetAllGroupsByGuru(ctx context.Context, guruId int, page int, size int, search string) ([]*entity.Group, int, error)
	GetGroupById(ctx context.Context, groupId string) (*entity.Group, error)
	UpdateGroup(ctx context.Context, group *entity.Group) (*entity.Group, error)
	DeleteGroup(ctx context.Context, groupId string) error

	// Group Members
	AddMember(ctx context.Context, groupId string, memberId int) error
	RemoveMember(ctx context.Context, groupId string, memberId int) error
	GetGroupMembers(ctx context.Context, groupId string) ([]*entity.GroupMember, error)
	IsMemberInGroup(ctx context.Context, groupId string, memberId int) (bool, error)

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
