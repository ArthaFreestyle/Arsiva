package repository

import (
	"ArthaFreestyle/Arsiva/internal/entity"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type UserRepository interface {
	FindByEmail(ctx context.Context,email string) (*entity.User, error)
	
}

type UserRepositoryImpl struct {
	DB *pgxpool.Pool
	Log *logrus.Logger
}

func NewUserRepository(db *pgxpool.Pool, log *logrus.Logger) UserRepository {
	return &UserRepositoryImpl{
		DB: db,
		Log: log,
	}
}


func (r *UserRepositoryImpl) FindByEmail(ctx context.Context,email string) (*entity.User, error) {
	SQL := `SELECT u.user_id, u.username, u.email, u.password_hash, u.role,u.created_at,u.last_login,u.is_active FROM users u
	WHERE u.email = $1`
	
	rows,err := r.DB.Query(context.Background(),SQL,email)
	if err != nil {
		return nil,err
	}

	user,err := pgx.CollectExactlyOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return user,nil
}
