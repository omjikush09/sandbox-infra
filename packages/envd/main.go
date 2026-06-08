package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/envd/router"
)

func main() {

	app := fiber.New()

	PORT := "3000"
	router.ExecuteRoute(app)
	if err := app.Listen(":" + PORT); err != nil {
		panic(err)
	}
}
