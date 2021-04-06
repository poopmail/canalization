package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func parseQueryInt(key string, fallback int, ctx *fiber.Ctx) (int, error) {
	value := ctx.Query(key, "")
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
