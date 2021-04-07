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
	router.Get("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.MiddlewareInjectRefreshToken, app.EndpointGetAccountRefreshToken)
	router.Patch("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.MiddlewareInjectRefreshToken, app.EndpointPatchAccountRefreshToken)
	router.Delete("/accounts/:identifier/refresh_tokens/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectAccount(true), app.EndpointDeleteAccountRefreshToken)

	router.Get("/mailboxes/check/:address", app.MiddlewareHandleBasicAuth, app.EndpointCheckMailboxAddress)
	router.Get("/mailboxes", app.MiddlewareHandleBasicAuth, app.EndpointGetMailboxes)
	router.Get("/mailboxes/:address", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectMailbox(true), app.EndpointGetMailbox)
	router.Post("/mailboxes", app.MiddlewareHandleBasicAuth, app.EndpointCreateMailbox)
	router.Delete("/mailboxes/:address", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectMailbox(true), app.EndpointDeleteMailbox)

	router.Get("/messages", app.MiddlewareHandleBasicAuth, app.EndpointGetMessages)
	router.Get("/messages/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectMessage(true), app.EndpointGetMessage)
	router.Delete("/messages/:id", app.MiddlewareHandleBasicAuth, app.MiddlewareInjectMessage(true), app.EndpointDeleteMessage)

	router.Get("/invites", app.MiddlewareHandleBasicAuth, app.MiddlewareRequireAdminAuth, app.EndpointGetInvites)
	router.Get("/invites/:code", app.MiddlewareHandleBasicAuth, app.MiddlewareRequireAdminAuth, app.MiddlewareInjectInvite, app.EndpointGetInvite)
	router.Post("/invites", app.MiddlewareHandleBasicAuth, app.MiddlewareRequireAdminAuth, app.EndpointCreateInvite)
	router.Delete("/invites/:code", app.MiddlewareHandleBasicAuth, app.MiddlewareRequireAdminAuth, app.MiddlewareInjectInvite, app.EndpointDeleteInvite)

	router.Post("/auth/refresh_token", app.EndpointPostRefreshToken)
	router.Get("/auth/access_token", app.EndpointGetAccessToken)
}
