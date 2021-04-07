package v1

import (
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/shared"
)

// MiddlewareInjectMailbox handles mailbox injection
func (app *App) MiddlewareInjectMailbox(handleAuth bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		claims := ctx.Locals("_claims").(*accessTokenClaims)

		address := ctx.Params("address")

		// Retrieve the requested mailbox
		mailbox, err := app.Mailboxes.Mailbox(address)
		if err != nil {
			return err
		}
		if mailbox == nil {
			return fiber.NewError(fiber.StatusNotFound, "mailbox not found")
		}

		// Handle authorization if needed
		if handleAuth && mailbox.Account != claims.ID && !claims.Admin {
			return fiber.ErrForbidden
		}

		ctx.Locals("_mailbox", mailbox)
		return ctx.Next()
	}
}

// EndpointCheckMailboxAddress handles the 'GET /v1/mailboxes/check/:address' API endpoint
func (app *App) EndpointCheckMailboxAddress(ctx *fiber.Ctx) error {
	address := ctx.Params("address")

	mailbox, err := app.Mailboxes.Mailbox(address)
	if err != nil {
		return err
	}

	return ctx.JSON(fiber.Map{
		"available": mailbox == nil,
	})
}

// EndpointGetMailboxes handles the 'GET /v1/mailboxes' API endpoint
func (app *App) EndpointGetMailboxes(ctx *fiber.Ctx) error {
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

	// Retrieve the account depending on the given 'account' query parameter
	var account *shared.Account
	accountName := ctx.Query("account")
	if strings.ToLower(accountName) == "@me" {
		found, err := app.Accounts.Account(claims.ID)
		if err != nil {
			return err
		}

		if found == nil {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}

		account = found
	} else if strings.HasPrefix(accountName, "@") {
		accountName = strings.TrimPrefix(accountName, "@")

		id, err := snowflake.ParseString(accountName)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
		}

		found, err := app.Accounts.Account(id)
		if err != nil {
			return err
		}

		if found == nil {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}

		account = found
	} else if accountName != "" {
		found, err := app.Accounts.AccountByUsername(accountName)
		if err != nil {
			return err
		}

		if found == nil {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}

		account = found
	}

	// Handle authentication
	if (account == nil && !claims.Admin) || (account.ID != claims.ID && !claims.Admin) {
		return fiber.ErrForbidden
	}

	// Retrieve the desired amount of mailboxes
	var count int
	var mailboxes []*shared.Mailbox
	if account == nil {
		count, err = app.Mailboxes.Count()
		if err != nil {
			return err
		}

		mailboxes, err = app.Mailboxes.Mailboxes(skip, limit)
		if err != nil {
			return err
		}
	} else {
		count, err = app.Mailboxes.CountInAccount(account.ID)
		if err != nil {
			return err
		}

		mailboxes, err = app.Mailboxes.MailboxesInAccount(account.ID, skip, limit)
		if err != nil {
			return err
		}
	}

	return ctx.JSON(newPaginatedResponse(mailboxes, count, len(mailboxes)))
}

// EndpointGetMailbox handles the 'GET /v1/mailboxes/:address' API endpoint
func (app *App) EndpointGetMailbox(ctx *fiber.Ctx) error {
	return ctx.JSON(ctx.Locals("_mailbox").(*shared.Mailbox))
}

// EndpointDeleteMailbox handles the 'DELETE /v1/mailboxes/:address' API endpoint
func (app *App) EndpointDeleteMailbox(ctx *fiber.Ctx) error {
	mailbox := ctx.Locals("_mailbox").(*shared.Mailbox)

	// Delete all messages of the mailbox
	if err := app.Messages.DeleteInMailbox(mailbox.Address); err != nil {
		return err
	}

	// Delete the mailbox itself
	return app.Mailboxes.Delete(mailbox.Address)
}
