package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/logging"
)

func TestCreateDemoSessionReturnsRepositorySession(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{
		session: domain.DemoSession{ID: "session-1", WalletID: "wallet-1"},
	}, fakeAppSessionRepository{}, logging.NewDiscardLogger("session_service_test"))

	session, err := service.CreateDemoSession(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID != "session-1" {
		t.Fatalf("expected session-1, got %s", session.ID)
	}
}

func TestAuthenticateRejectsExpiredSession(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{
		session: domain.DemoSession{
			ID:        "session-1",
			Token:     "token-1",
			WalletID:  "wallet-1",
			ExpiresAt: time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
		},
	}, fakeAppSessionRepository{}, logging.NewDiscardLogger("session_service_test"))
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	if _, err := service.Authenticate(context.Background(), "token-1"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestAuthenticateRejectsLookupErrors(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{
		err: errors.New("lookup failed"),
	}, fakeAppSessionRepository{}, logging.NewDiscardLogger("session_service_test"))

	if _, err := service.Authenticate(context.Background(), "token-1"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected ErrInvalidSession, got %v", err)
	}
}

func TestCreateRegisteredSessionReturnsRepositorySession(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{}, fakeAppSessionRepository{
		session: domain.AppSession{ID: "app-session-1", WalletID: "wallet-1"},
	}, logging.NewDiscardLogger("session_service_test"))

	session, err := service.CreateRegisteredSession(context.Background(), "user-1", "wallet-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if session.ID != "app-session-1" {
		t.Fatalf("expected app-session-1, got %s", session.ID)
	}
}

func TestAuthenticateRegisteredRejectsExpiredAppSession(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{}, fakeAppSessionRepository{
		session: domain.AppSession{
			ID:                "app-session-1",
			WalletID:          "wallet-1",
			IdleExpiresAt:     time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
			AbsoluteExpiresAt: time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
		},
	}, logging.NewDiscardLogger("session_service_test"))
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	if _, err := service.AuthenticateRegistered(context.Background(), "token-1"); !errors.Is(err, ErrInvalidAppSession) {
		t.Fatalf("expected ErrInvalidAppSession, got %v", err)
	}
}

func TestAuthenticateRegisteredRejectsRevokedSession(t *testing.T) {
	revokedAt := time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC)
	service := NewService(fakeDemoSessionRepository{}, fakeAppSessionRepository{
		session: domain.AppSession{
			ID:                "app-session-1",
			WalletID:          "wallet-1",
			IdleExpiresAt:     time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
			AbsoluteExpiresAt: time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
			RevokedAt:         &revokedAt,
		},
	}, logging.NewDiscardLogger("session_service_test"))

	if _, err := service.AuthenticateRegistered(context.Background(), "token-1"); !errors.Is(err, ErrInvalidAppSession) {
		t.Fatalf("expected ErrInvalidAppSession, got %v", err)
	}
}

func TestRevokeRegisteredSessionIsIdempotentForMissingToken(t *testing.T) {
	service := NewService(fakeDemoSessionRepository{}, fakeAppSessionRepository{}, logging.NewDiscardLogger("session_service_test"))

	if err := service.RevokeRegisteredSession(context.Background(), ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

type fakeDemoSessionRepository struct {
	session domain.DemoSession
	err     error
}

func (f fakeDemoSessionRepository) CreateDemoSession(context.Context) (domain.DemoSession, error) {
	return f.session, f.err
}

func (f fakeDemoSessionRepository) GetByToken(context.Context, string) (domain.DemoSession, error) {
	return f.session, f.err
}

type fakeAppSessionRepository struct {
	session   domain.AppSession
	err       error
	revoked   string
	revokeErr error
}

func (f fakeAppSessionRepository) CreateRegisteredSession(context.Context, string, string) (domain.AppSession, error) {
	return f.session, f.err
}

func (f fakeAppSessionRepository) GetRegisteredSessionByToken(context.Context, string) (domain.AppSession, error) {
	return f.session, f.err
}

func (f fakeAppSessionRepository) RevokeRegisteredSessionByToken(context.Context, string) error {
	return f.revokeErr
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}
