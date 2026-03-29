package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestAuthenticateReturnsErrorForMissingToken(t *testing.T) {
	service := NewService(fakeSessionRepository{})

	_, err := service.Authenticate(context.Background(), "")
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestAuthenticateReturnsErrorForExpiredSession(t *testing.T) {
	service := NewService(fakeSessionRepository{
		session: domain.DemoSession{
			ID:        "session-1",
			UserID:    "user-1",
			WalletID:  "wallet-1",
			ExpiresAt: time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC),
		},
	})
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	_, err := service.Authenticate(context.Background(), "token")
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestAuthenticateReturnsSessionForValidToken(t *testing.T) {
	expiresAt := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
	service := NewService(fakeSessionRepository{
		session: domain.DemoSession{
			ID:        "session-1",
			UserID:    "user-1",
			WalletID:  "wallet-1",
			ExpiresAt: expiresAt,
		},
	})
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	session, err := service.Authenticate(context.Background(), "token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.WalletID != "wallet-1" {
		t.Fatalf("expected wallet-1, got %s", session.WalletID)
	}
}

func TestCreateDemoSessionReturnsSession(t *testing.T) {
	service := NewService(fakeSessionRepository{
		session: domain.DemoSession{
			ID:       "session-1",
			UserID:   "user-1",
			WalletID: "wallet-1",
			Token:    "token",
		},
	})

	session, err := service.CreateDemoSession(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.Token != "token" {
		t.Fatalf("expected token to be returned")
	}
}

type fakeSessionRepository struct {
	session domain.DemoSession
	err     error
}

func (f fakeSessionRepository) CreateDemoSession(context.Context) (domain.DemoSession, error) {
	return f.session, f.err
}

func (f fakeSessionRepository) GetByToken(context.Context, string) (domain.DemoSession, error) {
	return f.session, f.err
}

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }
