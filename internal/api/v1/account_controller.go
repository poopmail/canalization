package v1

import (
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/hashing"
	"github.com/poopmail/canalization/internal/id"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/poopmail/canalization/internal/validation"
)

// MiddlewareInjectAccount handles account injection and authorization
func (app *App) MiddlewareInjectAccount(handleAuth bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		value := ctx.Params("identifier")
		claims := ctx.Locals("_claims").(*accessTokenClaims)

		var account *shared.Account
		var err error

		// Call the corresponding account retrieving function depending on the given parameter
		if strings.ToLower(value) == "@me" {
			account, err = app.Accounts.Account(claims.ID)
		} else if strings.HasPrefix(value, "@") {
			value = strings.TrimPrefix(value, "@")

			id, idErr := snowflake.ParseString(value)
			if idErr != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
			}

			account, err = app.Accounts.Account(id)
		} else {
			account, err = app.Accounts.AccountByUsername(value)
		}

		if err != nil {
			return err
		}

		if account == nil {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}

		if handleAuth && claims.ID != account.ID && !claims.Admin {
			return fiber.ErrForbidden
		}

		ctx.Locals("_account", account)
		return ctx.Next()
	}
}

// EndpointGetAccounts handles the 'GET /v1/accounts' API endpoint
func (app *App) EndpointGetAccounts(ctx *fiber.Ctx) error {
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

	// Count the total amount of accounts
	count, err := app.Accounts.Count()
	if err != nil {
		return err
	}

	// Retrieve the desired amount of accounts
	accounts, err := app.Accounts.Accounts(skip, limit)
	if err != nil {
		return err
	}

	// Remove the passwords from all retrieved accounts
	processed := make([]shared.Account, 0, len(accounts))
	for _, account := range accounts {
		copy := *account
		copy.Password = ""
		processed = append(processed, copy)
	}

	return ctx.JSON(newPaginatedResponse(processed, count, len(processed)))
}

// EndpointGetAccount handles the 'GET /v1/accounts/:identifier' API endpoint
func (app *App) EndpointGetAccount(ctx *fiber.Ctx) error {
	account := ctx.Locals("_account").(*shared.Account)
	copy := *account
	copy.Password = ""
	return ctx.JSON(copy)
}

type endpointCreateAccountRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Invite   string `json:"invite"`
}

// EndpointCreateAccount handles the 'POST /v1/accounts' API endpoint
func (app *App) EndpointCreateAccount(ctx *fiber.Ctx) error {
	// Try to parse the request into a request body struct
	body := new(endpointCreateAccountRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}
	if body.Username == "" || body.Password == "" || body.Invite == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bad request body")
	}

	// Validate the username syntax
	if !validation.ValidateAccountName(body.Username) {
		return fiber.NewError(fiber.StatusPreconditionFailed, "username violates restrictions")
	}

	// Check if an account with that username already exists
	found, err := app.Accounts.AccountByUsername(body.Username)
	if err != nil {
		return err
	}
	if found != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "username taken")
	}

	// Validate and delete the invite code
	invite, err := app.Invites.Invite(body.Invite)
	if err != nil {
		return err
	}
	if invite == nil {
		return fiber.NewError(fiber.StatusPreconditionFailed, "invalid invite")
	}
	if err := app.Invites.Delete(invite.Code); err != nil {
		return err
	}

	// Create the account
	hash, err := hashing.Hash(body.Password)
	if err != nil {
		return err
	}
	account := &shared.Account{
		ID:       id.Generate(),
		Username: body.Username,
		Password: hash,
		Admin:    false,
		Created:  time.Now().Unix(),
	}
	if err := app.Accounts.CreateOrReplace(account); err != nil {
		return err
	}
	copy := *account
	copy.Password = ""
	return ctx.Status(fiber.StatusCreated).JSON(copy)
}

type endpointPatchAccountRequestBody struct {
	Password string `json:"password"`
	Admin    *bool  `json:"admin"`
}

// EndpointPatchAccount handles the 'PATCH /v1/accounts/:identifier' API endpoint
func (app *App) EndpointPatchAccount(ctx *fiber.Ctx) error {
	// Try to parse the request into a request body struct
	body := new(endpointPatchAccountRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	// Check if the executor is an admin if the admin field should be changed
	if body.Admin != nil && !ctx.Locals("_claims").(*accessTokenClaims).Admin {
		return fiber.ErrForbidden
	}

	// Update the account
	account := ctx.Locals("_account").(*shared.Account)
	if body.Password != "" {
		hash, err := hashing.Hash(body.Password)
		if err != nil {
			return err
		}
		account.Password = hash
	}
	if body.Admin != nil {
		account.Admin = *body.Admin
	}
	if err := app.Accounts.CreateOrReplace(account); err != nil {
		return err
	}

	copy := *account
	copy.Password = ""
	return ctx.JSON(copy)
}

// EndpointDeleteAccount handles the 'DELETE /v1/accounts/:identifier' API endpoint
func (app *App) EndpointDeleteAccount(ctx *fiber.Ctx) error {
	account := ctx.Locals("_account").(*shared.Account)

	// Retrieve all mailboxes of the requested account and delete all messages in them
	mailboxAmount, err := app.Mailboxes.CountInAccount(account.ID)
	if err != nil {
		return err
	}
	mailboxes, err := app.Mailboxes.MailboxesInAccount(account.ID, 0, mailboxAmount)
	if err != nil {
		return err
	}
	for _, mailbox := range mailboxes {
		if err := app.Messages.DeleteInMailbox(mailbox.Address); err != nil {
			return err
		}
	}

	// Delete all mailboxes in the account themselves
	if err := app.Mailboxes.DeleteInAccount(account.ID); err != nil {
		return err
	}

	// Delete all refresh tokens of the account
	if err := app.RefreshTokens.DeleteAll(account.ID); err != nil {
		return err
	}

	// Delete the account
	return app.Accounts.Delete(account.ID)
}

// EndpointGetAccountRefreshTokens handles the 'GET /v1/accounts/:identifier/refresh_tokens' API endpoint
func (app *App) EndpointGetAccountRefreshTokens(ctx *fiber.Ctx) error {
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

	account := ctx.Locals("_account").(*shared.Account)

	// Count the total amount of refresh tokens
	count, err := app.RefreshTokens.Count(account.ID)
	if err != nil {
		return err
	}

	// Retrieve the desired amount of refresh tokens
	refreshTokens, err := app.RefreshTokens.RefreshTokens(account.ID, skip, limit)
	if err != nil {
		return err
	}

	// Remove the tokens from all retrieved refresh tokens
	processed := make([]shared.RefreshToken, 0, len(refreshTokens))
	for _, refreshToken := range refreshTokens {
		copy := *refreshToken
		copy.Token = ""
		processed = append(processed, copy)
	}

	return ctx.JSON(newPaginatedResponse(processed, count, len(processed)))
}

// EndpointGetAccountRefreshToken handles the 'GET /v1/accounts/:identifier/refresh_tokens/:id' API endpoint
func (app *App) EndpointGetAccountRefreshToken(ctx *fiber.Ctx) error {
	// Parse the snowflake ID of the refresh token
	rawID := ctx.Params("id")
	id, err := snowflake.ParseString(rawID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
	}

	account := ctx.Locals("_account").(*shared.Account)

	// Retrieve the refresh token
	refreshToken, err := app.RefreshTokens.RefreshToken(account.ID, id)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return fiber.NewError(fiber.StatusNotFound, "refresh token not found")
	}

	// Remove the token from the retrieved refresh token
	copy := *refreshToken
	copy.Token = ""
	return ctx.JSON(copy)
}

type endpointPatchAccountRefreshTokenRequestBody struct {
	Description *string `json:"description"`
}

// EndpointPatchAccountRefreshToken handles the 'PATCH /v1/accounts/:identifier/refresh_tokens/:id' API endpoint
func (app *App) EndpointPatchAccountRefreshToken(ctx *fiber.Ctx) error {
	// Parse the snowflake ID of the refresh token
	rawID := ctx.Params("id")
	id, err := snowflake.ParseString(rawID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
	}

	account := ctx.Locals("_account").(*shared.Account)

	// Retrieve the refresh token
	refreshToken, err := app.RefreshTokens.RefreshToken(account.ID, id)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return fiber.NewError(fiber.StatusNotFound, "refresh token not found")
	}

	// Try to parse the request into a request body struct
	body := new(endpointPatchAccountRefreshTokenRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}

	// Update the refresh token
	if body.Description != nil {
		refreshToken.Description = *body.Description
	}
	if err := app.RefreshTokens.CreateOrReplace(refreshToken); err != nil {
		return err
	}

	copy := *refreshToken
	copy.Token = ""
	return ctx.JSON(copy)
}

// EndpointDeleteAccountRefreshToken handles the 'DELETE /v1/accounts/:identifier/refresh_tokens/:id' API endpoint
func (app *App) EndpointDeleteAccountRefreshToken(ctx *fiber.Ctx) error {
	account := ctx.Locals("_account").(*shared.Account)

	rawID := ctx.Params("id")
	if strings.ToLower(rawID) != "@all" {
		// Parse the snowflake ID of the refresh token
		id, err := snowflake.ParseString(rawID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid snowflake ID")
		}

		// Delete the refresh token
		return app.RefreshTokens.Delete(account.ID, id)
	}

	// Delete all refresh tokens
	return app.RefreshTokens.DeleteAll(account.ID)
}
