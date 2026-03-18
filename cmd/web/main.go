package main

import (
	"ArthaFreestyle/Arsiva/internal/config"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func main() {

	viper := config.NewViper()
	app := config.NewFiber(viper)

	webPort := viper.GetInt("web.port")
	prefork := viper.GetBool("web.prefork")

	err := app.Listen(fmt.Sprintf(":%d", webPort),fiber.ListenConfig{
		EnablePrefork: prefork,
	})

	if err != nil {
		panic(err)
	}
}