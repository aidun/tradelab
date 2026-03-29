package account

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/aidun/tradelab/backend/internal/domain"
)

type ClerkVerifierConfig struct {
	JWKSURL  string
	Issuer   string
	MockMode bool
}

type ClerkTokenVerifier struct {
	keyfunc  keyfunc.Keyfunc
	issuer   string
	mockMode bool
}

type ClerkClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

func NewClerkTokenVerifier(ctx context.Context, cfg ClerkVerifierConfig) (*ClerkTokenVerifier, error) {
	if cfg.MockMode && (cfg.JWKSURL == "" || cfg.Issuer == "") {
		return &ClerkTokenVerifier{
			mockMode: true,
		}, nil
	}

	if cfg.JWKSURL == "" || cfg.Issuer == "" {
		return nil, nil
	}

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.JWKSURL})
	if err != nil {
		return nil, err
	}

	return &ClerkTokenVerifier{
		keyfunc:  jwks,
		issuer:   cfg.Issuer,
		mockMode: cfg.MockMode,
	}, nil
}

func (v *ClerkTokenVerifier) VerifyToken(ctx context.Context, token string) (domain.RegisteredIdentity, error) {
	if v == nil {
		return domain.RegisteredIdentity{}, ErrRegisteredAuthUnavailable
	}

	if v.mockMode && strings.HasPrefix(token, "mock-clerk:") {
		return parseMockIdentity(token)
	}

	if v.keyfunc == nil {
		return domain.RegisteredIdentity{}, ErrRegisteredAuthUnavailable
	}

	claims := ClerkClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
	)

	parsedToken, err := parser.ParseWithClaims(token, &claims, v.keyfunc.KeyfuncCtx(ctx))
	if err != nil {
		return domain.RegisteredIdentity{}, err
	}

	if !parsedToken.Valid {
		return domain.RegisteredIdentity{}, errors.New("token is not valid")
	}

	clerkUserID := strings.TrimSpace(claims.Subject)
	if clerkUserID == "" {
		return domain.RegisteredIdentity{}, fmt.Errorf("missing subject claim")
	}

	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = "Trader " + truncateID(clerkUserID)
	}

	return domain.RegisteredIdentity{
		ClerkUserID: clerkUserID,
		Email:       strings.TrimSpace(claims.Email),
		DisplayName: displayName,
	}, nil
}

func truncateID(value string) string {
	if len(value) <= 8 {
		return value
	}

	return value[:8]
}

func parseMockIdentity(token string) (domain.RegisteredIdentity, error) {
	parts := strings.Split(token, ":")
	if len(parts) < 3 {
		return domain.RegisteredIdentity{}, fmt.Errorf("invalid mock token")
	}

	provider := parts[1]
	slug := parts[2]
	displayName := "Mock " + strings.Title(provider) + " User"

	return domain.RegisteredIdentity{
		ClerkUserID: "mock_" + provider + "_" + slug,
		Email:       slug + "@" + provider + ".mock.tradelab",
		DisplayName: displayName,
	}, nil
}
