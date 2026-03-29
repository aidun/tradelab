package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	sessionservice "github.com/aidun/tradelab/backend/internal/service/session"
)

type MarketLister interface {
	List(ctx context.Context) ([]domain.Market, error)
}

type MarketCandlesLister interface {
	ListCandles(ctx context.Context, marketSymbol string, interval string, limit int) ([]domain.Candle, error)
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

func NewRouter(markets MarketLister, marketCandles MarketCandlesLister, orders OrderPlacer, portfolios PortfolioGetter, orderHistory OrderHistoryLister, activityHistory ActivityHistoryLister, sessions DemoSessionManager) http.Handler {
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

		candles, err := marketCandles.ListCandles(
			r.Context(),
			marketSymbol,
			r.URL.Query().Get("interval"),
			parseBoundedLimit(r.URL.Query().Get("limit"), 48, 1, 200),
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load candles")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"candles": candles})
	})

	mux.HandleFunc("POST /api/v1/sessions/demo", func(w http.ResponseWriter, r *http.Request) {
		demoSession, err := sessions.CreateDemoSession(r.Context())
		if err != nil {
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

	mux.HandleFunc("GET /api/v1/portfolios/", func(w http.ResponseWriter, r *http.Request) {
		demoSession, ok := requireSession(w, r, sessions)
		if !ok {
			return
		}

		walletID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolios/")
		if walletID == "" {
			writeError(w, http.StatusBadRequest, "wallet ID is required")
			return
		}

		if walletID != demoSession.WalletID {
			writeError(w, http.StatusForbidden, "wallet access denied")
			return
		}

		summary, err := portfolios.GetSummary(r.Context(), walletID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load portfolio")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"portfolio": summary})
	})

	mux.HandleFunc("GET /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		demoSession, ok := requireSession(w, r, sessions)
		if !ok {
			return
		}

		items, err := orderHistory.ListOrders(r.Context(), demoSession.WalletID, parseLimit(r.URL.Query().Get("limit")))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load orders")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"orders": items})
	})

	mux.HandleFunc("GET /api/v1/activity", func(w http.ResponseWriter, r *http.Request) {
		demoSession, ok := requireSession(w, r, sessions)
		if !ok {
			return
		}

		items, err := activityHistory.ListActivity(r.Context(), demoSession.WalletID, parseLimit(r.URL.Query().Get("limit")))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load activity")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"activity": items})
	})

	mux.HandleFunc("POST /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		demoSession, ok := requireSession(w, r, sessions)
		if !ok {
			return
		}

		var payload struct {
			MarketSymbol  string  `json:"market_symbol"`
			QuoteAmount   float64 `json:"quote_amount"`
			ExpectedPrice float64 `json:"expected_price"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		order, err := orders.PlaceMarketBuy(r.Context(), orderservice.PlaceMarketBuyInput{
			UserID:        demoSession.UserID,
			WalletID:      demoSession.WalletID,
			MarketSymbol:  payload.MarketSymbol,
			QuoteAmount:   payload.QuoteAmount,
			ExpectedPrice: payload.ExpectedPrice,
		})
		if err != nil {
			statusCode := http.StatusInternalServerError

			switch {
			case errors.Is(err, orderservice.ErrQuoteAmountTooLow), errors.Is(err, orderservice.ErrExpectedPriceLow):
				statusCode = http.StatusBadRequest
			case errors.Is(err, orderservice.ErrInsufficientFunds):
				statusCode = http.StatusUnprocessableEntity
			}

			writeError(w, statusCode, err.Error())
			return
		}

		writeJSON(w, http.StatusCreated, map[string]any{"order": order})
	})

	return mux
}

func requireSession(w http.ResponseWriter, r *http.Request, sessions DemoSessionManager) (domain.DemoSession, bool) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return domain.DemoSession{}, false
	}

	demoSession, err := sessions.Authenticate(r.Context(), token)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if !errors.Is(err, sessionservice.ErrInvalidSession) {
			statusCode = http.StatusInternalServerError
		}
		writeError(w, statusCode, "invalid session token")
		return domain.DemoSession{}, false
	}

	return demoSession, true
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
