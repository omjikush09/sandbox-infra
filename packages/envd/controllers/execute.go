package controllers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/sse"
	"github.com/omjikush09/sandboxing-infra/packages/envd/models"
	"github.com/omjikush09/sandboxing-infra/packages/envd/services"
)

func ExecuteJS(c fiber.Ctx) error {

	executeData := new(models.Execute)

	if err := c.Bind().Body(executeData); err != nil {
		return err
	}
	// fmt.Println(executeData)
	err := services.WriteToFile(executeData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	data := services.ExecuteJS()
	return c.Status(200).JSON(fiber.Map{
		"data":     data.Output,
		"output":   data.Output,
		"stdout":   data.Output,
		"stderr":   data.Stderr,
		"exitCode": data.ExitCode,
		"error":    data.Error,
	})
}

func ExecuteSync(c fiber.Ctx) error {

	executeCommand := models.ExecuteCommand{}

	if err := c.Bind().Body(executeCommand); err != nil {
		return fmt.Errorf("Failed to parse body")
	}

	stdout, stderr, err := services.ExecuteCmd(executeCommand.Command)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		stderr: stderr,
		stdout: stdout,
	})

}

func ExecuteAsync(c fiber.Ctx, stream *sse.Stream) error {
	executeCommand := models.ExecuteCommand{}

	if err := c.Bind().Body(executeCommand); err != nil {
		return fmt.Errorf("Failed to execute Command")
	}

	events, err := services.ExecuteCmdBackground(executeCommand.Command, stream.Context())

	if err != nil {
		return err
	}

	for {
		select {
		case msg, ok := <-events:
			if !ok {
				return nil
			}

			if err := stream.Event(sse.Event{Name: "message", Data: msg}); err != nil {
				return err
			}
		case <-stream.Done():
			return stream.Err()
		}
	}

}
