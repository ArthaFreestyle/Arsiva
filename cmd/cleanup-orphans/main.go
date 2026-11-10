package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ArthaFreestyle/Arsiva/internal/maintenance"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cleanup failed:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.json", "application config (DATABASE_URL overrides it)")
	uploadDir := flag.String("uploads", "./uploads", "uploads directory")
	minAge := flag.Duration("min-age", 7*24*time.Hour, "minimum file AND asset age")
	apply := flag.Bool("apply", false, "delete candidates; default is dry-run")
	offline := flag.Bool("offline", false, "confirm API, workers and other upload writers are stopped")
	flag.Parse()
	if *apply && !*offline {
		return fmt.Errorf("--apply requires --offline: stop API/workers/upload writers first")
	}
	dsn := os.Getenv("DATABASE_URL")
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("invalid database connection configuration")
	}
	if dsn == "" {
		v := viper.New()
		v.SetConfigFile(*configPath)
		if err := v.ReadInConfig(); err != nil {
			return err
		}
		cfg.Host = v.GetString("database.postgres.host")
		port := v.GetInt("database.postgres.port")
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid postgres port")
		}
		cfg.Port = uint16(port)
		cfg.User = v.GetString("database.postgres.user")
		cfg.Password = v.GetString("database.postgres.password")
		cfg.Database = v.GetString("database.postgres.dbname")
		cfg.TLSConfig = nil
		cfg.Fallbacks = nil
	}
	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("database connection failed (check config/DATABASE_URL and connectivity)")
	}
	defer conn.Close(ctx)
	return maintenance.Run(ctx, conn, maintenance.Options{UploadDir: *uploadDir, MinAge: *minAge, Apply: *apply, Output: os.Stdout})
}
