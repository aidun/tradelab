package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

var (
	ErrInvalidRegisteredToken    = errors.New("invalid registered authentication token")
	ErrRegisteredAuthUnavailable = errors.New("registered authentication is unavailable")
	ErrRegisteredAccountNotFound = errors.New("registered account not found")
	ErrGuestTokenRequired        = errors.New("guest token is required")
	ErrRegisteredAccountExists   = errors.New("registered account already exists")
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (domain.RegisteredIdentity, error)
}

type Service struct {
	accounts store.RegisteredAccountRepository
	verifier TokenVerifier
	logger   *slog.Logger
}

func NewService(accounts store.RegisteredAccountRepository, verifier TokenVerifier, logger *slog.Logger) *Service {
	return &Service{
		accounts: accounts,
		verifier: verifier,
		logger:   logger,
	}
}

func (s *Service) AuthenticateRegistered(ctx context.Context, token string) (domain.RegisteredAccount, error) {
	identity, err := s.verifyToken(ctx, token)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}

	account, err := s.accounts.GetByClerkUserID(ctx, identity.ClerkUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logInfo("authenticate_registered.not_bootstrapped", "clerk_user_id", identity.ClerkUserID)
			return domain.RegisteredAccount{}, ErrRegisteredAccountNotFound
		}

		s.logError("authenticate_registered.lookup_failed", err, "clerk_user_id", identity.ClerkUserID)
		return domain.RegisteredAccount{}, err
	}

	s.logInfo("authenticate_registered.success", "clerk_user_id", identity.ClerkUserID, "wallet_id", account.WalletID)
	return account, nil
}

func (s *Service) BootstrapRegisteredAccount(ctx context.Context, token string) (domain.RegisteredAccount, error) {
	identity, err := s.verifyToken(ctx, token)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}

	account, err := s.accounts.BootstrapRegisteredAccount(ctx, identity)
	if err != nil {
		s.logError("bootstrap_registered_account.failed", err, "clerk_user_id", identity.ClerkUserID)
		return domain.RegisteredAccount{}, err
	}

	s.logInfo("bootstrap_registered_account.success", "clerk_user_id", identity.ClerkUserID, "wallet_id", account.WalletID)
	return account, nil
}

func (s *Service) UpgradeGuestSession(ctx context.Context, registeredToken string, guestToken string, preserveGuestData bool) (domain.RegisteredAccount, error) {
	if guestToken == "" {
		return domain.RegisteredAccount{}, ErrGuestTokenRequired
	}

	identity, err := s.verifyToken(ctx, registeredToken)
	if err != nil {
		return domain.RegisteredAccount{}, err
	}

	account, err := s.accounts.UpgradeGuestSession(ctx, guestToken, identity, preserveGuestData)
	if err != nil {
		s.logError("upgrade_guest_session.failed", err, "clerk_user_id", identity.ClerkUserID, "preserve_guest_data", preserveGuestData)
		return domain.RegisteredAccount{}, err
	}

	s.logInfo("upgrade_guest_session.success", "clerk_user_id", identity.ClerkUserID, "wallet_id", account.WalletID, "preserve_guest_data", preserveGuestData)
	return account, nil
}

func (s *Service) verifyToken(ctx context.Context, token string) (domain.RegisteredIdentity, error) {
	if token == "" {
		return domain.RegisteredIdentity{}, ErrInvalidRegisteredToken
	}

	if s.verifier == nil {
		return domain.RegisteredIdentity{}, ErrRegisteredAuthUnavailable
	}

	identity, err := s.verifier.VerifyToken(ctx, token)
	if err != nil {
		s.logInfo("verify_registered_token.failed", "reason", err.Error())
		return domain.RegisteredIdentity{}, fmt.Errorf("%w: %v", ErrInvalidRegisteredToken, err)
	}

	return identity, nil
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
