package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/vm/controller"
)

func Start(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/execute", controller.Execute)
}
