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
	GetAllUsers(ctx context.Context) ([]*entity.User, error)
	GetUserById(ctx context.Context,userId string) (*entity.User, error)
	CreateUser(ctx context.Context,user *entity.User) (*entity.User, error)
	UpdateUser(ctx context.Context,user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context,user *entity.User) (error)
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
	r.Log.Info("query : ",SQL)

	user,err := pgx.CollectExactlyOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return user,nil
}

func (r *UserRepositoryImpl) GetAllUsers(ctx context.Context) ([]*entity.User, error) {
	SQL := `SELECT u.user_id, u.username, u.email, u.role FROM users u
	WHERE u.is_active = true`
	
	rows,err := r.DB.Query(context.Background(),SQL)
	if err != nil {
		return nil,err
	}

	r.Log.Info("query : ",SQL)
	users,err := pgx.CollectRows(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return users,nil
}

func (r *UserRepositoryImpl) GetUserById(ctx context.Context,userId string) (*entity.User, error) {
	SQL := `SELECT u.user_id, u.username, u.email, u.role FROM users u
	WHERE u.user_id = $1`
	
	rows,err := r.DB.Query(context.Background(),SQL,userId)
	if err != nil {
		return nil,err
	}

	r.Log.Info("query : ",SQL)
	user,err := pgx.CollectExactlyOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return user,nil
}

func (r *UserRepositoryImpl) CreateUser(ctx context.Context,user *entity.User) (*entity.User, error) {
	SQL := `INSERT INTO users (username,email,password_hash,role) VALUES ($1,$2,$3,$4) RETURNING user_id,username,email,role,created_at`
	
	rows,err := r.DB.Query(context.Background(),SQL,user.Username,user.Email,user.PasswordHash,user.Role)
	if err != nil {
		return nil,err
	}

	r.Log.Info("query : ",SQL)
	user,err = pgx.CollectExactlyOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return user,nil
}

func (r *UserRepositoryImpl) UpdateUser(ctx context.Context,user *entity.User) (*entity.User, error) {
	SQL := `UPDATE users SET username = $1,email = $2,password_hash = $3,role = $4 WHERE user_id = $5 RETURNING user_id,username,email,role,created_at`
	
	rows,err := r.DB.Query(context.Background(),SQL,user.Username,user.Email,user.PasswordHash,user.Role,user.UserId)
	if err != nil {
		return nil,err
	}

	r.Log.Info("query : ",SQL)
	user,err = pgx.CollectExactlyOneRow(rows,pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil,err
	}

	return user,nil
}

func (r *UserRepositoryImpl) DeleteUser(ctx context.Context,user *entity.User) (error) {
	SQL := `UPDATE users SET is_active = false WHERE user_id = $1`
	
	_,err := r.DB.Exec(context.Background(),SQL,user.UserId)
	if err != nil {
		return err
	}

	r.Log.Info("query : ",SQL)

	return nil
}
	
	