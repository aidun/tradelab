package session

import (
	"context"
	"errors"
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
	clock    Clock
}

func NewService(sessions store.DemoSessionRepository) *Service {
	return &Service{
		sessions: sessions,
		clock:    realClock{},
	}
}

func (s *Service) CreateDemoSession(ctx context.Context) (domain.DemoSession, error) {
	return s.sessions.CreateDemoSession(ctx)
}

func (s *Service) Authenticate(ctx context.Context, token string) (domain.DemoSession, error) {
	if token == "" {
		return domain.DemoSession{}, ErrInvalidSession
	}

	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return domain.DemoSession{}, ErrInvalidSession
	}

	if !session.ExpiresAt.After(s.clock.Now()) {
		return domain.DemoSession{}, ErrInvalidSession
	}

	return session, nil
}
