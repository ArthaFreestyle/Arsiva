package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogrus(config *viper.Viper) *logrus.Logger {
	logger := logrus.New()

	logger.SetLevel(logrus.Level(config.GetInt("log.level")))
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Default-default biar config lama yang cuma punya "log.level" tetep jalan
	config.SetDefault("log.file", "./logs/arsiva.log")
	config.SetDefault("log.max_size_mb", 50)
	config.SetDefault("log.max_backups", 7)
	config.SetDefault("log.max_age_days", 30)
	config.SetDefault("log.compress", true)

	logFile := config.GetString("log.file")

	// Lumberjack bikin file-nya tapi gak bikin folder parent-nya, jadi kita bikin sendiri.
	// Kalau gagal, jangan panic, cukup tetep pakai stderr aja.
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		logger.SetOutput(os.Stderr)
		logger.Warnf("Gagal bikin folder log %q, log cuma ke stderr: %v", filepath.Dir(logFile), err)
		return logger
	}

	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    config.GetInt("log.max_size_mb"),
		MaxBackups: config.GetInt("log.max_backups"),
		MaxAge:     config.GetInt("log.max_age_days"),
		Compress:   config.GetBool("log.compress"),
	}

	// Tulis ke file DAN stderr biar "docker compose logs" tetep kebaca
	multi := io.MultiWriter(os.Stderr, fileWriter)
	logger.SetOutput(multi)

	return logger
}
