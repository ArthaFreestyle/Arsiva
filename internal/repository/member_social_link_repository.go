package repository

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"ArthaFreestyle/Arsiva/internal/entity"
)

type MemberSocialLinkRepository interface {
	Create(ctx context.Context, link *entity.MemberSocialLink) (*entity.MemberSocialLink, error)
	FindById(ctx context.Context, socialId string) (*entity.MemberSocialLink, error)
	FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberSocialLink, error)
	Update(ctx context.Context, link *entity.MemberSocialLink) (*entity.MemberSocialLink, error)
	Delete(ctx context.Context, socialId string) error
	ExistsByMemberAndPlatform(ctx context.Context, memberId, platform string, excludeSocialId string) (bool, error)
}

type memberSocialLinkRepositoryImpl struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

func NewMemberSocialLinkRepository(db *pgxpool.Pool, log *logrus.Logger) MemberSocialLinkRepository {
	return &memberSocialLinkRepositoryImpl{DB: db, Log: log}
}

func (r *memberSocialLinkRepositoryImpl) Create(ctx context.Context, link *entity.MemberSocialLink) (*entity.MemberSocialLink, error) {
	memberId, err := strconv.Atoi(link.MemberId)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO member_social_links (member_id, platform, url)
		VALUES ($1, $2, $3)
		RETURNING social_id::text,
		          member_id::text,
		          platform::text,
		          url,
		          created_at::text
	`
	rows, err := r.DB.Query(ctx, query, memberId, link.Platform, link.URL)
	if err != nil {
		r.Log.Errorf("Error Create social link: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberSocialLink])
	if err != nil {
		r.Log.Errorf("Error collecting row Create social link: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberSocialLinkRepositoryImpl) FindById(ctx context.Context, socialId string) (*entity.MemberSocialLink, error) {
	id, err := strconv.Atoi(socialId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT social_id::text,
		       member_id::text,
		       platform::text,
		       url,
		       created_at::text
		FROM member_social_links
		WHERE social_id = $1
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error FindById social link: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberSocialLink])
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *memberSocialLinkRepositoryImpl) FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberSocialLink, error) {
	id, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT social_id::text,
		       member_id::text,
		       platform::text,
		       url,
		       created_at::text
		FROM member_social_links
		WHERE member_id = $1
		ORDER BY social_id ASC
	`
	rows, err := r.DB.Query(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error FindAllByMemberId social link: %v", err)
		return nil, err
	}
	defer rows.Close()

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberSocialLink])
	if err != nil {
		r.Log.Errorf("Error collecting rows FindAllByMemberId social link: %v", err)
		return nil, err
	}
	return links, nil
}

func (r *memberSocialLinkRepositoryImpl) Update(ctx context.Context, link *entity.MemberSocialLink) (*entity.MemberSocialLink, error) {
	id, err := strconv.Atoi(link.SocialId)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE member_social_links
		SET platform = $1,
		    url = $2
		WHERE social_id = $3
		RETURNING social_id::text,
		          member_id::text,
		          platform::text,
		          url,
		          created_at::text
	`
	rows, err := r.DB.Query(ctx, query, link.Platform, link.URL, id)
	if err != nil {
		r.Log.Errorf("Error Update social link: %v", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByNameLax[entity.MemberSocialLink])
	if err != nil {
		r.Log.Errorf("Error collecting row Update social link: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *memberSocialLinkRepositoryImpl) Delete(ctx context.Context, socialId string) error {
	id, err := strconv.Atoi(socialId)
	if err != nil {
		return err
	}

	query := `DELETE FROM member_social_links WHERE social_id = $1`
	_, err = r.DB.Exec(ctx, query, id)
	if err != nil {
		r.Log.Errorf("Error Delete social link: %v", err)
		return err
	}
	return nil
}

func (r *memberSocialLinkRepositoryImpl) ExistsByMemberAndPlatform(ctx context.Context, memberId, platform string, excludeSocialId string) (bool, error) {
	mid, err := strconv.Atoi(memberId)
	if err != nil {
		return false, err
	}

	var (
		exists bool
		query  string
	)

	if excludeSocialId == "" {
		query = `SELECT EXISTS(SELECT 1 FROM member_social_links WHERE member_id = $1 AND platform = $2)`
		err = r.DB.QueryRow(ctx, query, mid, platform).Scan(&exists)
	} else {
		eid, convErr := strconv.Atoi(excludeSocialId)
		if convErr != nil {
			return false, convErr
		}
		query = `SELECT EXISTS(SELECT 1 FROM member_social_links WHERE member_id = $1 AND platform = $2 AND social_id != $3)`
		err = r.DB.QueryRow(ctx, query, mid, platform, eid).Scan(&exists)
	}

	if err != nil {
		r.Log.Errorf("Error ExistsByMemberAndPlatform: %v", err)
		return false, err
	}
	return exists, nil
}
