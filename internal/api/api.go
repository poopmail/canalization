package api

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	recov "github.com/gofiber/fiber/v2/middleware/recover"
	v1 "github.com/poopmail/canalization/internal/api/v1"
	"github.com/poopmail/canalization/internal/config"
	"github.com/poopmail/canalization/internal/karen"
	"github.com/poopmail/canalization/internal/shared"
	"github.com/poopmail/canalization/internal/static"
	"github.com/sirupsen/logrus"
	"github.com/ztrue/tracerr"
)

// API represents an instance of the REST API
type API struct {
	app      *fiber.App
	Services *Services
}

// Services holds all services used by the REST API
type Services struct {
	Accounts      shared.AccountService
	RefreshTokens shared.RefreshTokenService
	Invites       shared.InviteService
	Mailboxes     shared.MailboxService
	Messages      shared.MessageService
	Redis         *redis.Client
}

// Serve serves the REST API
func (api *API) Serve() error {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(_ *fiber.Ctx, err error) error {
			if fiberErr, ok := err.(*fiber.Error); ok {
				if fiberErr.Code >= 500 {
					return karen.Send(api.Services.Redis, karen.Message{
						Type:        karen.MessageTypeError,
						Service:     static.KarenServiceName,
						Topic:       "API Request",
						Description: tracerr.Sprint(tracerr.Wrap(err)),
					})
				}
			}
			return nil
		},
		DisableKeepalive:      true,
		DisableStartupMessage: static.Production,
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
	if !static.Production {
		app.Use(logger.New())
		app.Use(pprof.New())
	}

	// Inject the rate limiter middleware
	app.Use(limiter.New(limiter.Config{
		Next: func(_ *fiber.Ctx) bool {
			return !static.Production
		},
		Max: config.Loaded.APIRateLimit,
		LimitReached: func(ctx *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}))

	// Route the v1 API endpoints
	(&v1.App{
		Accounts:      api.Services.Accounts,
		RefreshTokens: api.Services.RefreshTokens,
		Invites:       api.Services.Invites,
		Mailboxes:     api.Services.Mailboxes,
		Messages:      api.Services.Messages,
		Redis:         api.Services.Redis,
	}).Route(app.Group("/v1"))

	logrus.WithField("address", config.Loaded.APIAddress).Info("Serving the REST API")
	api.app = app
	return app.Listen(config.Loaded.APIAddress)
}

// Shutdown gracefully shuts down the REST API
func (api *API) Shutdown() error {
	logrus.Info("Shutting down the REST API")
	return api.app.Shutdown()
}
