package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
	userdata "github.com/omjikush09/sandboxing-infra/packages/vm/userData"
	pool "github.com/omjikush09/sandboxing-infra/packages/vm/vmpool"
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

func CreateUser(c fiber.Ctx) error {

	vm, ok := pool.GetInstance()
	if ok == false {
		return c.Status(500).JSON(fiber.Map{
			"error": "We are out of capacity",
		})
	}

	user := userdata.CreateUser()
	user.VM = vm
	return c.JSON(fiber.Map{
		"id": user.ID,
	})
}

func OpenPort(c fiber.Ctx) error {

	userId := c.Params("userId")
	port := c.Query("port")

	if port == "" {
		return fmt.Errorf("Port Id not found")
	}
	if userId == "" {
		return fmt.Errorf("User Id not found")
	}

	user, err := userdata.GetUser(userId)
	if err != nil {
		return err
	}
	user.Port = port
	return c.JSON(fiber.Map{
		"data": "Opened the PORT",
	})

}
