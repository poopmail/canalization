package v1

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/auth"
	"github.com/poopmail/canalization/internal/hashing"
	"github.com/poopmail/canalization/internal/shared"
)

// MiddlewareAuth handles the first stage of user authentication and injects the user as executor into the context
func MiddlewareAuth(adminRequired bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Try to parse the authorization header
		header := strings.SplitN(ctx.Get("Authorization"), " ", 2)
		if len(header) != 2 || header[0] != "Bearer" {
			return fiber.ErrUnauthorized
		}

		// Validate and extract the account out of the token
		authenticator := ctx.Locals("__services_authenticator").(auth.Authenticator)
		valid, account, err := authenticator.ProcessToken(header[1])
		if err != nil {
			return err
		}
		if !valid {
			return fiber.ErrUnauthorized
		}

		// Validate the admin state if required
		if adminRequired && !account.Admin {
			return fiber.ErrForbidden
		}

		ctx.Locals("_executor", account)
		return ctx.Next()
	}
}

type endpointPostAuthRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// EndpointPostAuth handles the 'POST /v1/auth' API endpoint
func EndpointPostAuth(ctx *fiber.Ctx) error {
	// Try to parse the request body into a struct
	body := new(endpointPostAuthRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}
	if body.Username == "" || body.Password == "" {
		return fiber.ErrBadRequest
	}

	// Try to retrieve the account the user wants to log in with
	accounts := ctx.Locals("__services_accounts").(shared.AccountService)
	account, err := accounts.AccountByUsername(body.Username)
	if err != nil {
		return err
	}
	if account == nil {
		return fiber.ErrUnauthorized
	}

	// Validate the given password
	valid, err := hashing.Check(body.Password, account.Password)
	if err != nil {
		return err
	}
	if !valid {
		return fiber.ErrUnauthorized
	}

	// Create and return a new authentication token
	authenticator := ctx.Locals("__services_authenticator").(auth.Authenticator)
	token, err := authenticator.GenerateToken(account)
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{
		"token": token,
	})
}
