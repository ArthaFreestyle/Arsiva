package config

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func NewRedis(config *viper.Viper) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr : config.GetString("database.redis.host") + ":" + config.GetString("database.redis.port"),
		Password : config.GetString("database.redis.password"),
		DB : config.GetInt("database.redis.db"),
	})

	return redisClient
}