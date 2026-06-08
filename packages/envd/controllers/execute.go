package controllers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/envd/models"
	"github.com/omjikush09/sandboxing-infra/packages/envd/services"
)

func ExecuteJS(c fiber.Ctx) error {

	executeData := new(models.Execute)

	if err := c.Bind().Body(executeData); err != nil {
		return err
	}
	// fmt.Println(executeData)
	data, err := services.ExecuteJS(executeData)
	if err != nil {
		c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	c.Status(200).JSON(fiber.Map{
		"data": data,
	})
	return nil
}
