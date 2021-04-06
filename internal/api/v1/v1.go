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
	// TODO: Domains

	router.Post("/auth/refresh_token", app.EndpointPostRefreshToken)
	router.Get("/auth/access_token", app.EndpointGetAccessToken)
}
