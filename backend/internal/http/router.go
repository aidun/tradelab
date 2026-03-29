package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	sessionservice "github.com/aidun/tradelab/backend/internal/service/session"
)

type MarketLister interface {
	List(ctx context.Context) ([]domain.Market, error)
}

type MarketCandlesLister interface {
	ListCandles(ctx context.Context, marketSymbol string, interval string, limit int) (domain.CandleFeed, error)
}

type OrderPlacer interface {
	PlaceMarketBuy(ctx context.Context, input orderservice.PlaceMarketBuyInput) (domain.Order, error)
}

type PortfolioGetter interface {
	GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error)
}

type OrderHistoryLister interface {
	ListOrders(ctx context.Context, walletID string, limit int) ([]domain.Order, error)
}

type ActivityHistoryLister interface {
	ListActivity(ctx context.Context, walletID string, limit int) ([]domain.ActivityLog, error)
}

type DemoSessionManager interface {
	CreateDemoSession(ctx context.Context) (domain.DemoSession, error)
	Authenticate(ctx context.Context, token string) (domain.DemoSession, error)
}

type RegisteredAccountManager interface {
	AuthenticateRegistered(ctx context.Context, token string) (domain.RegisteredAccount, error)
	BootstrapRegisteredAccount(ctx context.Context, token string) (domain.RegisteredAccount, error)
	UpgradeGuestSession(ctx context.Context, registeredToken string, guestToken string, preserveGuestData bool) (domain.RegisteredAccount, error)
}

func NewRouter(markets MarketLister, marketCandles MarketCandlesLister, orders OrderPlacer, portfolios PortfolioGetter, orderHistory OrderHistoryLister, activityHistory ActivityHistoryLister, sessions DemoSessionManager, registeredAccounts RegisteredAccountManager, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"name":   "tradelab-api",
		})
	})

	mux.HandleFunc("GET /api/v1/markets", func(w http.ResponseWriter, r *http.Request) {
		items, err := markets.List(r.Context())
		if err != nil {
			logError(logger, "markets.list_failed", err, "path", r.URL.Path)
			writeError(w, http.StatusInternalServerError, "failed to load markets")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"markets": items})
	})

	mux.HandleFunc("GET /api/v1/markets/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/candles") {
			writeError(w, http.StatusNotFound, "route not found")
			return
		}

		marketSymbol := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/markets/"), "/candles")
		if marketSymbol == "" {
			writeError(w, http.StatusBadRequest, "market symbol is required")
			return
		}

		marketSymbol, err := url.PathUnescape(marketSymbol)
		if err != nil {
			writeError(w, http.StatusBadRequest, "market symbol is invalid")
			return
		}

		feed, err := marketCandles.ListCandles(
			r.Context(),
			marketSymbol,
			r.URL.Query().Get("interval"),
			parseBoundedLimit(r.URL.Query().Get("limit"), 48, 1, 200),
		)
		if err != nil {
			logError(logger, "markets.candles_failed", err, "path", r.URL.Path, "market_symbol", marketSymbol)
			writeError(w, http.StatusInternalServerError, "failed to load candles")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"candles": feed.Candles,
			"meta":    feed.Meta,
		})
	})

	mux.HandleFunc("POST /api/v1/sessions/demo", func(w http.ResponseWriter, r *http.Request) {
		demoSession, err := sessions.CreateDemoSession(r.Context())
		if err != nil {
			logError(logger, "sessions.demo_create_failed", err, "path", r.URL.Path)
			writeError(w, http.StatusInternalServerError, "failed to create demo session")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{
			"session": map[string]any{
				"id":         demoSession.ID,
				"user_id":    demoSession.UserID,
				"wallet_id":  demoSession.WalletID,
				"token":      demoSession.Token,
				"expires_at": demoSession.ExpiresAt,
			},
		})
	})

	mux.HandleFunc("POST /api/v1/account/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			logInfo(logger, "auth.registered_missing_bearer_token", "path", r.URL.Path)
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		account, err := registeredAccounts.BootstrapRegisteredAccount(r.Context(), token)
		if err != nil {
			logError(logger, "account.bootstrap_failed", err, "path", r.URL.Path)
			writeError(w, http.StatusUnauthorized, "failed to bootstrap registered account")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"account": map[string]any{
				"user_id":       account.UserID,
				"wallet_id":     account.WalletID,
				"clerk_user_id": account.ClerkUserID,
				"email":         account.Email,
				"display_name":  account.DisplayName,
				"mode":          "registered",
			},
		})
	})

	mux.HandleFunc("POST /api/v1/account/upgrade", func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			logInfo(logger, "auth.registered_missing_bearer_token", "path", r.URL.Path)
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		guestToken := strings.TrimSpace(r.Header.Get("X-TradeLab-Guest-Token"))
		if guestToken == "" {
			writeError(w, http.StatusBadRequest, "guest token is required")
			return
		}

		var payload struct {
			PreserveGuestData bool `json:"preserve_guest_data"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		account, err := registeredAccounts.UpgradeGuestSession(r.Context(), token, guestToken, payload.PreserveGuestData)
		if err != nil {
			statusCode := http.StatusUnauthorized
			if strings.Contains(err.Error(), "already exists") {
				statusCode = http.StatusConflict
			}
			logError(logger, "account.upgrade_failed", err, "path", r.URL.Path, "status_code", statusCode)
			writeError(w, statusCode, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"account": map[string]any{
				"user_id":       account.UserID,
				"wallet_id":     account.WalletID,
				"clerk_user_id": account.ClerkUserID,
				"email":         account.Email,
				"display_name":  account.DisplayName,
				"mode":          "registered",
			},
		})
	})

	mux.HandleFunc("GET /api/v1/portfolios/", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, sessions, registeredAccounts, logger)
		if !ok {
			return
		}

		walletID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolios/")
		if walletID == "" {
			writeError(w, http.StatusBadRequest, "wallet ID is required")
			return
		}

		if walletID != principal.WalletID {
			writeError(w, http.StatusForbidden, "wallet access denied")
			return
		}

		summary, err := portfolios.GetSummary(r.Context(), walletID)
		if err != nil {
			logError(logger, "portfolios.summary_failed", err, "wallet_id", walletID)
			writeError(w, http.StatusInternalServerError, "failed to load portfolio")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"portfolio": summary})
	})

	mux.HandleFunc("GET /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, sessions, registeredAccounts, logger)
		if !ok {
			return
		}

		items, err := orderHistory.ListOrders(r.Context(), principal.WalletID, parseLimit(r.URL.Query().Get("limit")))
		if err != nil {
			logError(logger, "orders.history_failed", err, "wallet_id", principal.WalletID)
			writeError(w, http.StatusInternalServerError, "failed to load orders")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"orders": items})
	})

	mux.HandleFunc("GET /api/v1/activity", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, sessions, registeredAccounts, logger)
		if !ok {
			return
		}

		items, err := activityHistory.ListActivity(r.Context(), principal.WalletID, parseLimit(r.URL.Query().Get("limit")))
		if err != nil {
			logError(logger, "activity.history_failed", err, "wallet_id", principal.WalletID)
			writeError(w, http.StatusInternalServerError, "failed to load activity")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"activity": items})
	})

	mux.HandleFunc("POST /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePrincipal(w, r, sessions, registeredAccounts, logger)
		if !ok {
			return
		}

		var payload struct {
			MarketSymbol string  `json:"market_symbol"`
			QuoteAmount  float64 `json:"quote_amount"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logInfo(logger, "orders.create_invalid_payload", "path", r.URL.Path)
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		logInfo(logger, "orders.create_attempt", "wallet_id", principal.WalletID, "market_symbol", payload.MarketSymbol, "quote_amount", payload.QuoteAmount)
		order, err := orders.PlaceMarketBuy(r.Context(), orderservice.PlaceMarketBuyInput{
			UserID:       principal.UserID,
			WalletID:     principal.WalletID,
			MarketSymbol: payload.MarketSymbol,
			QuoteAmount:  payload.QuoteAmount,
		})
		if err != nil {
			statusCode := http.StatusInternalServerError

			switch {
			case errors.Is(err, orderservice.ErrQuoteAmountTooLow), errors.Is(err, orderservice.ErrCurrentPriceUnavailable):
				statusCode = http.StatusBadRequest
			case errors.Is(err, orderservice.ErrInsufficientFunds):
				statusCode = http.StatusUnprocessableEntity
			}

			logError(logger, "orders.create_failed", err, "wallet_id", principal.WalletID, "market_symbol", payload.MarketSymbol, "status_code", statusCode)
			writeError(w, statusCode, err.Error())
			return
		}

		logInfo(logger, "orders.create_success", "wallet_id", principal.WalletID, "market_symbol", order.MarketSymbol, "order_id", order.ID)
		writeJSON(w, http.StatusCreated, map[string]any{"order": order})
	})

	return loggingMiddleware(logger, mux)
}

func requirePrincipal(w http.ResponseWriter, r *http.Request, sessions DemoSessionManager, registeredAccounts RegisteredAccountManager, logger *slog.Logger) (domain.Principal, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		logInfo(logger, "auth.missing_bearer_token", "path", r.URL.Path)
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return domain.Principal{}, false
	}

	demoSession, err := sessions.Authenticate(r.Context(), token)
	if err == nil {
		return domain.Principal{
			Kind:      domain.PrincipalKindGuest,
			UserID:    demoSession.UserID,
			WalletID:  demoSession.WalletID,
			SessionID: demoSession.ID,
		}, true
	}

	if !errors.Is(err, sessionservice.ErrInvalidSession) {
		logError(logger, "auth.invalid_session", err, "path", r.URL.Path, "status_code", http.StatusInternalServerError)
		writeError(w, http.StatusInternalServerError, "invalid session token")
		return domain.Principal{}, false
	}

	// Guest sessions stay the default path, but once that lookup says "not a guest token"
	// we fall through to the registered-account verifier so both identity models can coexist.
	account, registeredErr := registeredAccounts.AuthenticateRegistered(r.Context(), token)
	if registeredErr != nil {
		logError(logger, "auth.invalid_registered_account", registeredErr, "path", r.URL.Path, "status_code", http.StatusUnauthorized)
		writeError(w, http.StatusUnauthorized, "invalid session token")
		return domain.Principal{}, false
	}

	return domain.Principal{
		Kind:        domain.PrincipalKindRegistered,
		UserID:      account.UserID,
		WalletID:    account.WalletID,
		ClerkUserID: account.ClerkUserID,
	}, true
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}

	return token, true
}

func parseLimit(raw string) int {
	return parseBoundedLimit(raw, 10, 1, 100)
}

func parseBoundedLimit(raw string, fallback int, min int, max int) int {
	if raw == "" {
		return fallback
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < min || limit > max {
		return fallback
	}
	return limit
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		// Request logs are the shared operational breadcrumb for route handlers and auth failures.
		logger.Info("request.started", "operation", "request.started", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(recorder, r)
		logger.Info(
			"request.completed",
			"operation", "request.completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", recorder.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func logInfo(logger *slog.Logger, operation string, args ...any) {
	if logger == nil {
		return
	}

	logger.Info(operation, append([]any{"operation", operation}, args...)...)
}

func logError(logger *slog.Logger, operation string, err error, args ...any) {
	if logger == nil {
		return
	}

	logger.Error(operation, append([]any{"operation", operation, "error", err}, args...)...)
}
