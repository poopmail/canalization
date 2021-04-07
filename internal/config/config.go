package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/poopmail/canalization/internal/env"
	"github.com/poopmail/canalization/internal/random"
)

// Loaded holds the currently loaded configuration object
var Loaded *Config

// Config represents the application configuration data
type Config struct {
	KarenRedisChannel           string
	PostgresDSN                 string
	RefreshTokenLifetime        time.Duration
	RefreshTokenCleanupInterval time.Duration
	AccessTokenLifetime         time.Duration
	AccessTokenSigningKey       []byte
	RedisURL                    string
	DomainOverride              []string
	APIAddress                  string
	APIRateLimit                int
	AccountMailboxLimit         int
}

func init() {
	godotenv.Load()

	Loaded = &Config{
		KarenRedisChannel:           env.MustString("CANAL_KAREN_REDIS_CHANNEL", "karen"),
		PostgresDSN:                 env.MustString("CANAL_POSTGRES_DSN", ""),
		RefreshTokenLifetime:        env.MustDuration("CANAL_REFRESH_TOKEN_LIFETIME", false, 7*24*time.Hour),
		RefreshTokenCleanupInterval: env.MustDuration("CANAL_REFRESH_TOKEN_CLEANUP_INTERVAL", false, 60*time.Minute),
		AccessTokenLifetime:         env.MustDuration("CANAL_ACCESS_TOKEN_LIFETIME", false, 15*time.Minute),
		AccessTokenSigningKey:       []byte(env.MustString("CANAL_ACCESS_TOKEN_SIGNING_KEY", random.RandomString(64))),
		RedisURL:                    env.MustString("CANAL_REDIS_URL", "redis://localhost:6379/0"),
		DomainOverride:              env.MustStringSlice("CANAL_DOMAIN_OVERRIDE", ",", []string{}),
		APIAddress:                  env.MustString("CANAL_API_ADDRESS", ":8080"),
		APIRateLimit:                env.MustInt("CANAL_API_RATE_LIMIT", 60),
		AccountMailboxLimit:         env.MustInt("CANAL_ACCOUNT_MAILBOX_LIMIT", 10),
	}
}
