package v1

import (
	"context"

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

// EndpointGetInfo handles the 'GET /v1/domains' API endpoint
func (app *App) EndpointGetDomains(ctx *fiber.Ctx) error {
	domains, err := app.Redis.SMembers(context.Background(), static.DomainsRedisKey).Result()
	if err != nil {
		return err
	}
	return ctx.JSON(domains)
}
