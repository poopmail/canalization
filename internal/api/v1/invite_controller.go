package v1

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/random"
	"github.com/poopmail/canalization/internal/shared"
)

// MiddlewareInjectInvite handles invite injection
func (app *App) MiddlewareInjectInvite(ctx *fiber.Ctx) error {
	invite, err := app.Invites.Invite(ctx.Params("code"))
	if err != nil {
		return err
	}
	if invite == nil {
		return fiber.NewError(fiber.StatusNotFound, "invite not found")
	}

	ctx.Locals("_invite", invite)
	return ctx.Next()
}

// EndpointGetInvites handles the 'GET /v1/invites' API endpoint
func (app *App) EndpointGetInvites(ctx *fiber.Ctx) error {
	// Parse the 'skip' query parameter
	skip, err := parseQueryInt("skip", 0, ctx)
	if err != nil || skip < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "bad query parameter")
	}

	// Parse the 'limit' query parameter
	limit, err := parseQueryInt("limit", 10, ctx)
	if err != nil || limit < 0 {
		return fiber.NewError(fiber.StatusBadRequest, "bad query parameter")
	}

	// Retrieve the total amount of invites
	count, err := app.Invites.Count()
	if err != nil {
		return err
	}

	// Retrieve the desired amount of invites
	invites, err := app.Invites.Invites(skip, limit)
	if err != nil {
		return err
	}

	return ctx.JSON(newPaginatedResponse(invites, count, len(invites)))
}

// EndpointGetInvite handles the 'GET /v1/invites/:code' API endpoint
func (app *App) EndpointGetInvite(ctx *fiber.Ctx) error {
	return ctx.JSON(ctx.Locals("_invite").(*shared.Invite))
}

type endpointCreateInviteRequestBody struct {
	Code string `json:"string"`
}

// EndpointCreateInvite handles the 'POST /v1/invites' API endpoint
func (app *App) EndpointCreateInvite(ctx *fiber.Ctx) error {
	// Try to parse the request into a request body struct
	body := new(endpointCreateInviteRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	// Define the code the invite should have
	code := body.Code
	if code == "" {
		code = random.RandomString(32)
	}

	// Create the invite
	invite := &shared.Invite{
		Code:    code,
		Created: time.Now().Unix(),
	}
	if err := app.Invites.CreateOrReplace(invite); err != nil {
		return err
	}

	return ctx.JSON(invite)
}

// EndpointDeleteInvite handles the 'DELETE /v1/invites/:code' API endpoint
func (app *App) EndpointDeleteInvite(ctx *fiber.Ctx) error {
	invite := ctx.Locals("_invite").(*shared.Invite)
	return app.Invites.Delete(invite.Code)
}
