package v1

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/poopmail/canalization/internal/config"
	"github.com/poopmail/canalization/internal/hashing"
	"github.com/poopmail/canalization/internal/id"
	"github.com/poopmail/canalization/internal/random"
	"github.com/poopmail/canalization/internal/shared"
)

// MiddlewareHandleBasicAuth handles basic access token validation
func (app *App) MiddlewareHandleBasicAuth(ctx *fiber.Ctx) error {
	header := strings.SplitN(ctx.Get(fiber.HeaderAuthorization), " ", 2)
	if len(header) != 2 || header[0] != "Bearer" {
		return fiber.ErrUnauthorized
	}

	valid, claims, _ := app.processAccessToken(header[1])
	if !valid {
		return fiber.ErrUnauthorized
	}

	ctx.Locals("_claims", claims)
	return ctx.Next()
}

// MiddlewareRequireAdminAuth requires admin authentication
func (app *App) MiddlewareRequireAdminAuth(ctx *fiber.Ctx) error {
	if !ctx.Locals("_claims").(*accessTokenClaims).Admin {
		return fiber.ErrForbidden
	}
	return ctx.Next()
}

type endpointPostRefreshTokenRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// EndpointPostRefreshToken handles the 'POST /v1/auth/refresh_token' API endpoint
func (app *App) EndpointPostRefreshToken(ctx *fiber.Ctx) error {
	// Try to parse the request into a request body struct
	body := new(endpointPostRefreshTokenRequestBody)
	if err := ctx.BodyParser(body); err != nil {
		return err
	}
	if body.Username == "" || body.Password == "" {
		return fiber.ErrBadRequest
	}

	// Try to retrieve the account the client is trying to authenticate with
	account, err := app.Accounts.AccountByUsername(body.Username)
	if err != nil {
		return err
	}
	if account == nil {
		return fiber.ErrUnauthorized
	}

	// Validate the given password
	if valid, _ := hashing.Check(body.Password, account.Password); !valid {
		return fiber.ErrUnauthorized
	}

	// Generate and create a new refresh token
	rawToken := random.RandomString(64)
	hashedToken, err := hashing.Hash(rawToken)
	if err != nil {
		return err
	}
	token := &shared.RefreshToken{
		ID:          id.Generate(),
		Account:     account.ID,
		Token:       hashedToken,
		Description: "",
		Created:     time.Now().Unix(),
	}
	if err := app.RefreshTokens.CreateOrReplace(token); err != nil {
		return err
	}

	// Set the cookie on the clients side
	ctx.Cookie(&fiber.Cookie{
		Name:     "_refresh_token",
		Value:    base64.StdEncoding.EncodeToString([]byte(account.ID.String() + ":" + rawToken)),
		Path:     "/v1/auth/access_token",
		Expires:  time.Now().Add(config.Loaded.RefreshTokenLifetime),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
	return ctx.SendStatus(fiber.StatusOK)
}

// EndpointGetAccessToken handles the 'GET /v1/auth/access_token' API endpoint
func (app *App) EndpointGetAccessToken(ctx *fiber.Ctx) error {
	// Try to read the refresh token cookie
	refreshTokenValue := ctx.Cookies("_refresh_token")
	if refreshTokenValue == "" {
		return fiber.ErrUnauthorized
	}

	// Try to decode the value present in the refresh token cookie
	decoded, err := base64.StdEncoding.DecodeString(refreshTokenValue)
	if err != nil {
		return fiber.ErrUnauthorized
	}

	split := strings.SplitN(string(decoded), ":", 2)

	// Try to retrieve the account mentioned in the cookie
	accountID, err := snowflake.ParseString(split[0])
	if err != nil {
		return fiber.ErrUnauthorized
	}
	account, err := app.Accounts.Account(accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return fiber.ErrUnauthorized
	}

	// Retrieve all refresh tokens from that account
	amount, err := app.RefreshTokens.Count(accountID)
	if err != nil {
		return err
	}
	refreshTokens, err := app.RefreshTokens.RefreshTokens(accountID, 0, amount)
	if err != nil {
		return err
	}

	// Loop through all these refresh tokens and compare them to the given one
	var refreshToken *shared.RefreshToken
	for _, potentialRefreshToken := range refreshTokens {
		if potentialRefreshToken.Created < time.Now().Add(-config.Loaded.RefreshTokenLifetime).Unix() {
			continue
		}

		if valid, _ := hashing.Check(split[1], potentialRefreshToken.Token); valid {
			refreshToken = potentialRefreshToken
			break
		}
	}
	if refreshToken == nil {
		return fiber.ErrUnauthorized
	}

	// Issue a new access token
	expires := time.Now().Add(config.Loaded.AccessTokenLifetime).Unix()
	accessToken, err := app.issueAccessToken(account, expires)
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{
		"access_token": accessToken,
		"expires":      expires,
	})
}

type accessTokenClaims struct {
	jwt.StandardClaims
	ID    snowflake.ID `json:"c_id"`
	Admin bool         `json:"c_admin"`
}

func (app *App) issueAccessToken(account *shared.Account, expires int64) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires,
			IssuedAt:  time.Now().Unix(),
			Subject:   account.ID.String(),
		},
		ID:    account.ID,
		Admin: account.Admin,
	}).SignedString(config.Loaded.AccessTokenSigningKey)
}

func (app *App) processAccessToken(token string) (bool, *accessTokenClaims, error) {
	claims := new(accessTokenClaims)
	parsed, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) {
		return config.Loaded.AccessTokenSigningKey, nil
	})
	if err != nil {
		return false, nil, err
	}

	if !parsed.Valid {
		return false, nil, nil
	}

	return true, claims, nil
}
