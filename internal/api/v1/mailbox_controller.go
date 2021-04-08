package v1

import (
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/config"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/poopmail/canalization/internal/static"
	"github.com/poopmail/canalization/internal/validation"
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
	if (account == nil && !claims.Admin) || (account != nil && account.ID != claims.ID && !claims.Admin) {
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

type endpointCreateMailboxRequestBody struct {
	Key     string `json:"key"`
	Domain  string `json:"domain"`
	Account string `json:"account"`
}

// EndpointCreateMailbox handles the 'POST /v1/mailboxes' API endpoint
func (app *App) EndpointCreateMailbox(ctx *fiber.Ctx) error {
	// Try to parse the request into a request body struct
	body := new(endpointCreateMailboxRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}
	if body.Key == "" || body.Domain == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bad request body")
	}

	accountName := "@me"
	if body.Account != "" {
		accountName = strings.ToLower(body.Account)
	}

	claims := ctx.Locals("_claims").(*accessTokenClaims)

	// Retrieve the account
	var account *shared.Account
	var err error
	if accountName == "@me" {
		account, err = app.Accounts.Account(claims.ID)
	} else {
		account, err = app.Accounts.AccountByUsername(accountName)
	}
	if err != nil {
		return err
	}
	if account == nil {
		return fiber.NewError(fiber.StatusNotFound, "account not found")
	}

	// Handle authorization
	if account.ID != claims.ID && !claims.Admin {
		return fiber.ErrForbidden
	}

	// Validate the mailbox key
	if !validation.ValidateMailboxKey(body.Key) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid mailbox key")
	}

	// Validate the mailbox domain
	isValidDomain, err := app.Redis.SIsMember(ctx.Context(), static.DomainsRedisKey, strings.ToLower(body.Domain)).Result()
	if err != nil {
		return err
	}
	if !isValidDomain {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid mailbox domain")
	}

	// Check if the account has exceeded its mailbox limit
	if !claims.Admin && !account.Admin {
		count, err := app.Mailboxes.CountInAccount(account.ID)
		if err != nil {
			return err
		}

		if count >= config.Loaded.AccountMailboxLimit {
			return fiber.NewError(fiber.StatusPreconditionFailed, "mailbox limit exceeded")
		}
	}

	// Create the mailbox
	address := body.Key + "@" + body.Domain
	mailbox := &shared.Mailbox{
		Address: address,
		Account: account.ID,
		Created: time.Now().Unix(),
	}
	if err := app.Mailboxes.CreateOrReplace(mailbox); err != nil {
		return err
	}

	return ctx.JSON(mailbox)
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
