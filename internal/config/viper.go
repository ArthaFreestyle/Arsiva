package config

import (
	"github.com/spf13/viper"
	"fmt"
)

func NewViper() *viper.Viper{
	config := viper.New()

	config.SetConfigName("config")
	config.SetConfigType("json")
	config.AddConfigPath("../../")
	config.AddConfigPath("./")

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))	
	}

	return config
}