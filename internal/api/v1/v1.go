package v1

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/auth"
	"github.com/poopmail/canalization/internal/shared"
)

// App represents the v1 API app
type App struct {
	Address           string
	RequestsPerMinute int
	Production        bool
	Version           string

	Authenticator auth.Authenticator
	Invites       shared.InviteService
	Accounts      shared.AccountService
	Mailboxes     shared.MailboxService
	Messages      shared.MessageService
	Redis         *redis.Client
}

// Route routes the v1 API endpoints
func (app *App) Route(router fiber.Router) {
	router.Get("/info", app.EndpointGetInfo)
	router.Get("/domains", app.MiddlewareAuth(false), app.EndpointGetDomains)

	router.Post("/auth", app.EndpointPostAuth)
}
