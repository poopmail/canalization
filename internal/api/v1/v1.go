package v1

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/shared"
)

// App represents the v1 API app
type App struct {
	Accounts      shared.AccountService
	RefreshTokens shared.RefreshTokenService
	Invites       shared.InviteService
	Mailboxes     shared.MailboxService
	Messages      shared.MessageService
	Redis         *redis.Client
}

// Route routes the v1 API endpoints
func (app *App) Route(router fiber.Router) {
	router.Get("/info", app.EndpointGetInfo)
	router.Get("/domains", app.MiddlewareHandleBasicAuth, app.EndpointGetDomains)

	router.Get("/accounts", app.MiddlewareHandleBasicAuth, app.EndpointGetAccounts)
	router.Get("/accounts/:identifier", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointGetAccount)
	router.Post("/accounts", app.EndpointCreateAccount)
	router.Patch("/accounts/:identifier", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointPatchAccount)
	router.Delete("/accounts/:identifier", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointDeleteAccount)
	router.Get("/accounts/:identifier/refresh_tokens", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointGetAccountRefreshTokens)
	router.Get("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointGetAccountRefreshToken)
	router.Patch("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointPatchAccountRefreshToken)
	router.Delete("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointDeleteAccountRefreshToken)

	router.Post("/auth/refresh_token", app.EndpointPostRefreshToken)
	router.Get("/auth/access_token", app.EndpointGetAccessToken)
}
