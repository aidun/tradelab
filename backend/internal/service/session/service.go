package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/logging"
	"github.com/aidun/tradelab/backend/internal/store"
)

var ErrInvalidSession = errors.New("invalid session token")
var ErrInvalidAppSession = errors.New("invalid app session")

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type Service struct {
	sessions    store.DemoSessionRepository
	appSessions store.AppSessionRepository
	logger      *slog.Logger
	clock       Clock
}

func NewService(sessions store.DemoSessionRepository, appSessions store.AppSessionRepository, logger *slog.Logger) *Service {
	return &Service{
		sessions:    sessions,
		appSessions: appSessions,
		logger:      logger,
		clock:       realClock{},
	}
}

func (s *Service) CreateDemoSession(ctx context.Context) (domain.DemoSession, error) {
	session, err := s.sessions.CreateDemoSession(ctx)
	if err != nil {
		s.logError("create_demo_session.failed", err)
		return domain.DemoSession{}, err
	}

	s.logInfo("create_demo_session.success", "session_id", session.ID, "wallet_id", session.WalletID)
	return session, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (domain.DemoSession, error) {
	if token == "" {
		s.logInfo("authenticate.invalid_token", "reason", "missing")
		return domain.DemoSession{}, ErrInvalidSession
	}

	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		s.logInfo("authenticate.invalid_token", "reason", "lookup_failed")
		return domain.DemoSession{}, ErrInvalidSession
	}

	if !session.ExpiresAt.After(s.clock.Now()) {
		s.logInfo("authenticate.invalid_token", "session_id", session.ID, "reason", "expired")
		return domain.DemoSession{}, ErrInvalidSession
	}

	s.logInfo("authenticate.success", "session_id", session.ID, "wallet_id", session.WalletID)
	return session, nil
}

func (s *Service) CreateRegisteredSession(ctx context.Context, userID string, walletID string) (domain.AppSession, error) {
	if s.appSessions == nil {
		return domain.AppSession{}, ErrInvalidAppSession
	}

	session, err := s.appSessions.CreateRegisteredSession(ctx, userID, walletID)
	if err != nil {
		s.logError("create_registered_session.failed", err, "user_id", userID, "wallet_id", walletID)
		return domain.AppSession{}, err
	}

	s.logInfo("create_registered_session.success", "session_id", session.ID, "wallet_id", session.WalletID)
	return session, nil
}

func (s *Service) AuthenticateRegistered(ctx context.Context, token string) (domain.AppSession, error) {
	if token == "" || s.appSessions == nil {
		return domain.AppSession{}, ErrInvalidAppSession
	}

	session, err := s.appSessions.GetRegisteredSessionByToken(ctx, token)
	if err != nil {
		s.logInfo("authenticate_registered_session.invalid", "reason", "lookup_failed")
		return domain.AppSession{}, ErrInvalidAppSession
	}

	now := s.clock.Now()
	if session.RevokedAt != nil || !session.IdleExpiresAt.After(now) || !session.AbsoluteExpiresAt.After(now) {
		s.logInfo("authenticate_registered_session.invalid", "session_id", session.ID, "reason", "expired_or_revoked")
		return domain.AppSession{}, ErrInvalidAppSession
	}

	s.logInfo("authenticate_registered_session.success", "session_id", session.ID, "wallet_id", session.WalletID)
	return session, nil
}

func (s *Service) RevokeRegisteredSession(ctx context.Context, token string) error {
	if token == "" || s.appSessions == nil {
		return nil
	}

	if err := s.appSessions.RevokeRegisteredSessionByToken(ctx, token); err != nil {
		s.logError("revoke_registered_session.failed", err)
		return err
	}

	s.logInfo("revoke_registered_session.success")
	return nil
}

func (s *Service) logInfo(operation string, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Info(operation, append([]any{"operation", operation}, args...)...)
}

func (s *Service) logError(operation string, err error, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Error(operation, append([]any{"operation", operation, "error", logging.RedactError(err)}, args...)...)
}
