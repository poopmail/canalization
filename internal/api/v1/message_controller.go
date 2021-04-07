package v1

import (
	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/shared"
)

// MiddlewareInjectMessage handles message injection
func (app *App) MiddlewareInjectMessage(handleAuth bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		rawID := ctx.Params("id")
		id, err := snowflake.ParseString(rawID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
		}

		// Retrieve the requested message
		message, err := app.Messages.Message(id)
		if err != nil {
			return err
		}
		if message == nil {
			return fiber.NewError(fiber.StatusNotFound, "message not found")
		}

		// Retrieve the corresponding mailbox
		mailbox, err := app.Mailboxes.Mailbox(message.Mailbox)
		if err != nil {
			return err
		}
		if mailbox == nil {
			return fiber.NewError(fiber.StatusInternalServerError, "mailbox mapped but not present")
		}

		// Handle authorization if needed
		claims := ctx.Locals("_claims").(*accessTokenClaims)
		if handleAuth && mailbox.Account != claims.ID && !claims.Admin {
			return fiber.ErrForbidden
		}

		ctx.Locals("_message", message)
		return ctx.Next()
	}
}

// EndpointGetMessages handles the 'GET /v1/messages' API endpoint
func (app *App) EndpointGetMessages(ctx *fiber.Ctx) error {
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

	claims := ctx.Locals("_claims").(*accessTokenClaims)

	// Retrieve the mailbox depending on the given 'mailbox' query parameter
	mailboxAddress := ctx.Query("mailbox")
	if mailboxAddress == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bad query parameter")
	}
	mailbox, err := app.Mailboxes.Mailbox(mailboxAddress)
	if err != nil {
		return err
	}
	if mailbox == nil {
		return fiber.NewError(fiber.StatusNotFound, "mailbox not found")
	}

	// Handle authentication
	if mailbox.Account != claims.ID && !claims.Admin {
		return fiber.ErrForbidden
	}

	// Retrieve the desired amount of messages
	count, err := app.Messages.Count(mailbox.Address)
	if err != nil {
		return err
	}
	messages, err := app.Messages.Messages(mailbox.Address, skip, limit)
	if err != nil {
		return err
	}

	return ctx.JSON(newPaginatedResponse(messages, count, len(messages)))
}

// EndpointGetMessage handles the 'GET /v1/messages/:id' API endpoint
func (app *App) EndpointGetMessage(ctx *fiber.Ctx) error {
	return ctx.JSON(ctx.Locals("_message").(*shared.Message))
}

// EndpointDeleteMessage handles the 'DELETE /v1/messages/:id' API endpoint
func (app *App) EndpointDeleteMessage(ctx *fiber.Ctx) error {
	message := ctx.Locals("_message").(*shared.Message)
	return app.Messages.Delete(message.ID)
}
