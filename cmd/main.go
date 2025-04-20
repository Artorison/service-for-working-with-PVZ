package main

import (
	"log"
	"os"
	"pvz/internal/app"
	"pvz/internal/config"
	"pvz/internal/validation"

	"github.com/labstack/echo/v4"
)

const ConfigDefault = "config/config.yml"

func main() {

	router := echo.New()
	router.Validator = validation.NewValidator()

	filepath := os.Getenv("CFG_FILEPATH")
	if filepath == "" {
		filepath = ConfigDefault
	}

	cfg, err := config.LoadConfig(filepath)
	if err != nil {
		log.Fatal(err)
	}

	ap, err := app.NewApp(router, *cfg)
	if err != nil {
		log.Fatal(err)
	}
	ap.RegisterRoutes()
	ap.RegisterMiddlewares()
	ap.Start()
}
