package session

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

var ErrInvalidSession = errors.New("invalid session token")

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type Service struct {
	sessions store.DemoSessionRepository
	logger   *slog.Logger
	clock    Clock
}

func NewService(sessions store.DemoSessionRepository, logger *slog.Logger) *Service {
	return &Service{
		sessions: sessions,
		logger:   logger,
		clock:    realClock{},
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

	s.logger.Error(operation, append([]any{"operation", operation, "error", err}, args...)...)
}
