package controller

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/omjikush09/sandboxing-infra/packages/vm/pool"
)

func Execute(c fiber.Ctx) error {

	vm, ok := pool.GetInstance()
	if ok == false {
		return c.Status(500).JSON(fiber.Map{
			"error": "We are out of capacity",
		})
	}
	defer pool.DeleteVM(vm)

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
