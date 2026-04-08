package main

import (
	"ArthaFreestyle/Arsiva/internal/config"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func main() {

	viper := config.NewViper()
	app := config.NewFiber(viper)
	db, err := config.NewPgx(viper)
	if err != nil {
		panic(err)
	}
	log := config.NewLogrus(viper)
	validate := config.NewValidator(viper)
	redis := config.NewRedis(viper)
	secret := []byte(viper.GetString("app.jwt-secret"))
	webPort := viper.GetInt("web.port")
	prefork := viper.GetBool("web.prefork")

	config.Bootstrap(config.BootstrapConfig{
		DB:       db,
		Redis:    redis,
		App:      app,
		Log:      log,
		Validate: validate,
		Secret:   secret,
		Config:   viper,
	})

	err = app.Listen(fmt.Sprintf(":%d", webPort), fiber.ListenConfig{
		EnablePrefork: prefork,
	})

	if err != nil {
		panic(err)
	}
}
