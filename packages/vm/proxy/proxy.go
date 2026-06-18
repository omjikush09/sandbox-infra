package proxy

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	userdata "github.com/omjikush09/sandboxing-infra/packages/vm/userData"
)

func getSubDomain(host string) string {

	subdomain := strings.Split(host, ".")
	if len(subdomain) < 2 {
		return ""
	}
	return subdomain[0]
}

func ProxyMiddleware(c fiber.Ctx) error {

	host := c.Hostname()
	subDomain := getSubDomain(host)

	if subDomain == "" {
		return c.Next()
	}

	user, err := userdata.GetUser(subDomain)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	if user.Port == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Port not open",
		})
	}
	target := fmt.Sprintf("http://%s,%s", user.VM.GuestIP, user.Port)

	return proxy.Do(c, target)

}
