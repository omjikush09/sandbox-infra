package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/envd/controllers"
)

func ExecuteRoute(app *fiber.App) {
	api := app.Group("/api")

	execute := api.Group("/execute")

	execute.Post("/js", controllers.ExecuteJS)

}
