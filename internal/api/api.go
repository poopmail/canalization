package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	recov "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/sirupsen/logrus"
)

// API represents an instance of the REST API
type API struct {
	app      *fiber.App
	Settings *Settings
	Services *Services
}

// Settings represents the settings used by the REST API
type Settings struct {
	Address           string
	RequestsPerMinute int
	Production        bool
	Version           string
}

// Services holds all services used by the REST API
type Services struct {
	Invites   shared.InviteService
	Accounts  shared.AccountService
	Mailboxes shared.MailboxService
	Messages  shared.MessageService
}

// Serve serves the REST API
func (api *API) Serve() error {
	app := fiber.New(fiber.Config{
		DisableKeepalive:      true,
		DisableStartupMessage: api.Settings.Production,
	})

	// Include CORS response headers
	app.Use(cors.New(cors.Config{
		Next:             nil,
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "",
		AllowCredentials: true,
		ExposeHeaders:    "",
		MaxAge:           0,
	}))

	// Enable panic recovering
	app.Use(recov.New())

	// Inject debug middlewares if the application runs in development mode
	if !api.Settings.Production {
		app.Use(logger.New())
		app.Use(pprof.New())
	}

	// Inject the rate limiter middleware
	app.Use(limiter.New(limiter.Config{
		Next: func(_ *fiber.Ctx) bool {
			return !api.Settings.Production
		},
		Max: api.Settings.RequestsPerMinute,
		LimitReached: func(ctx *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	logrus.WithField("address", api.Settings.Address).Info("Serving the REST API")
	api.app = app
	return app.Listen(api.Settings.Address)
}

// Shutdown gracefully shuts down the REST API
func (api *API) Shutdown() error {
	logrus.Info("Shutting down the REST API")
	return api.app.Shutdown()
}
