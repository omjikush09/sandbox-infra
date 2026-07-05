package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/envd/controllers"
	"github.com/gofiber/fiber/v3/middleware/sse"
)

func ExecuteRoute(app *fiber.App) {
	api := app.Group("/api")

	execute := api.Group("/execute")

	execute.Post("/js", controllers.ExecuteJS)

}

func ExecuteCommandSync(app *fiber.App) {
	api := app.Group("/api")

	execute := api.Group("/execute")

	// execute.Post()
	execute.Post("/sync", controllers.ExecuteSync)
}

func ExecuteCommandAsync(app *fiber.App) {
	api := app.Group("/api")

	execute := api.Group("/execute")
	
	execute.Post("/async", sse.New(sse.Config{
		Handler: controllers.ExecuteAsync,
	}))
}
