package account

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestBootstrapRegisteredAccountCreatesNewAccount(t *testing.T) {
	service := NewService(
		fakeAccountRepository{
			bootstrapAccount: domain.RegisteredAccount{
				UserID:      "user-1",
				WalletID:    "wallet-1",
				ClerkUserID: "clerk-1",
				Email:       "trader@example.com",
				DisplayName: "Trader Example",
			},
		},
		fakeTokenVerifier{
			identity: domain.RegisteredIdentity{
				ClerkUserID: "clerk-1",
				Email:       "trader@example.com",
				DisplayName: "Trader Example",
			},
		},
		testLogger(),
	)

	account, err := service.BootstrapRegisteredAccount(context.Background(), "registered-token")
	if err != nil {
		t.Fatalf("expected bootstrap to succeed, got %v", err)
	}

	if account.ClerkUserID != "clerk-1" {
		t.Fatalf("expected clerk user id to be returned, got %s", account.ClerkUserID)
	}
}

func TestAuthenticateRegisteredReturnsAccountForExistingUser(t *testing.T) {
	service := NewService(
		fakeAccountRepository{
			account: domain.RegisteredAccount{
				UserID:      "user-1",
				WalletID:    "wallet-1",
				ClerkUserID: "clerk-1",
			},
		},
		fakeTokenVerifier{
			identity: domain.RegisteredIdentity{ClerkUserID: "clerk-1"},
		},
		testLogger(),
	)

	account, err := service.AuthenticateRegistered(context.Background(), "registered-token")
	if err != nil {
		t.Fatalf("expected authenticate to succeed, got %v", err)
	}

	if account.WalletID != "wallet-1" {
		t.Fatalf("expected wallet-1, got %s", account.WalletID)
	}
}

func TestAuthenticateRegisteredReturnsNotFoundWhenAccountMissing(t *testing.T) {
	service := NewService(
		fakeAccountRepository{getErr: sql.ErrNoRows},
		fakeTokenVerifier{identity: domain.RegisteredIdentity{ClerkUserID: "clerk-1"}},
		testLogger(),
	)

	_, err := service.AuthenticateRegistered(context.Background(), "registered-token")
	if !errors.Is(err, ErrRegisteredAccountNotFound) {
		t.Fatalf("expected ErrRegisteredAccountNotFound, got %v", err)
	}
}

func TestUpgradeGuestSessionPreservesGuestData(t *testing.T) {
	service := NewService(
		fakeAccountRepository{
			upgradeAccount: domain.RegisteredAccount{
				UserID:      "user-registered",
				WalletID:    "wallet-guest",
				ClerkUserID: "clerk-1",
			},
		},
		fakeTokenVerifier{identity: domain.RegisteredIdentity{ClerkUserID: "clerk-1"}},
		testLogger(),
	)

	account, err := service.UpgradeGuestSession(context.Background(), "registered-token", "guest-token", true)
	if err != nil {
		t.Fatalf("expected upgrade to succeed, got %v", err)
	}

	if account.WalletID != "wallet-guest" {
		t.Fatalf("expected guest wallet to be preserved, got %s", account.WalletID)
	}
}

func TestUpgradeGuestSessionAllowsFreshRegisteredAccount(t *testing.T) {
	service := NewService(
		fakeAccountRepository{
			upgradeAccount: domain.RegisteredAccount{
				UserID:      "user-registered",
				WalletID:    "wallet-registered",
				ClerkUserID: "clerk-1",
			},
		},
		fakeTokenVerifier{identity: domain.RegisteredIdentity{ClerkUserID: "clerk-1"}},
		testLogger(),
	)

	account, err := service.UpgradeGuestSession(context.Background(), "registered-token", "guest-token", false)
	if err != nil {
		t.Fatalf("expected fresh upgrade to succeed, got %v", err)
	}

	if account.WalletID != "wallet-registered" {
		t.Fatalf("expected fresh registered wallet, got %s", account.WalletID)
	}
}

func TestUpgradeGuestSessionRequiresGuestToken(t *testing.T) {
	service := NewService(fakeAccountRepository{}, fakeTokenVerifier{}, testLogger())

	_, err := service.UpgradeGuestSession(context.Background(), "registered-token", "", true)
	if !errors.Is(err, ErrGuestTokenRequired) {
		t.Fatalf("expected ErrGuestTokenRequired, got %v", err)
	}
}

func TestAuthenticateRegisteredRejectsInvalidToken(t *testing.T) {
	service := NewService(
		fakeAccountRepository{},
		fakeTokenVerifier{err: errors.New("invalid token")},
		testLogger(),
	)

	_, err := service.AuthenticateRegistered(context.Background(), "registered-token")
	if !errors.Is(err, ErrInvalidRegisteredToken) {
		t.Fatalf("expected ErrInvalidRegisteredToken, got %v", err)
	}
}

type fakeAccountRepository struct {
	account          domain.RegisteredAccount
	bootstrapAccount domain.RegisteredAccount
	upgradeAccount   domain.RegisteredAccount
	getErr           error
	bootstrapErr     error
	upgradeErr       error
}

func (f fakeAccountRepository) GetByClerkUserID(context.Context, string) (domain.RegisteredAccount, error) {
	return f.account, f.getErr
}

func (f fakeAccountRepository) BootstrapRegisteredAccount(context.Context, domain.RegisteredIdentity) (domain.RegisteredAccount, error) {
	return f.bootstrapAccount, f.bootstrapErr
}

func (f fakeAccountRepository) UpgradeGuestSession(context.Context, string, domain.RegisteredIdentity, bool) (domain.RegisteredAccount, error) {
	return f.upgradeAccount, f.upgradeErr
}

type fakeTokenVerifier struct {
	identity domain.RegisteredIdentity
	err      error
}

func (f fakeTokenVerifier) VerifyToken(context.Context, string) (domain.RegisteredIdentity, error) {
	return f.identity, f.err
}
