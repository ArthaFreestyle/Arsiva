package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

func NewPgx(config *viper.Viper) (*pgxpool.Pool, error) {
	PostgresPort := config.GetString("database.postgres.port")
	PostgresHost := config.GetString("database.postgres.host")
	PostgresUser := config.GetString("database.postgres.user")
	PostgresPassword := config.GetString("database.postgres.password")
	PostgresDBName := config.GetString("database.postgres.dbname")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
		PostgresUser, PostgresPassword, PostgresHost, PostgresPort, PostgresDBName)

	poolConfig,err := pgxpool.ParseConfig(connStr)
	if err != nil {
		panic(fmt.Sprintf("Unable to parse connection string: %v\n", err))
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 3
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 10 * time.Minute
	
	ctx := context.Background()
	pool,err := pgxpool.NewWithConfig(ctx,poolConfig)
	if err != nil {
		panic(fmt.Sprintf("Unable to create connection pool: %v\n", err))
	}


	return pool,nil
}