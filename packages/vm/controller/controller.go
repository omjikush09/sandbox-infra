package controller

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/vm/start"
)

func Execute(c fiber.Ctx) error {

	vm, cmd, err := start.CreateVm()
	if err != nil {
		log.Printf("failed to create vm: %v", err)
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create vm",
		})
	}
	defer vm.Cleanup(cmd)

	resp, err := http.Post("http://"+vm.GuestIP+":3000/api/execute/js", "application/json", bytes.NewBuffer(c.Body()))

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to execute request in vm",
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to read vm response",
		})
	}

	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	return c.Status(resp.StatusCode).Send(body)
}
