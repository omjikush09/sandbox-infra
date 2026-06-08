package controller

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/vm/start"
)

func Execute(c fiber.Ctx) error {

	vm, cmd, err := start.CreateVm()
	if err != nil {

		println(err.Error())
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create vm",
		})
	}

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
