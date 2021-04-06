package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/static"
)

// EndpointGetInfo handles the 'GET /v1/info' API endpoint
func (app *App) EndpointGetInfo(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"production": static.Production,
		"version":    static.ApplicationVersion,
	})
}
