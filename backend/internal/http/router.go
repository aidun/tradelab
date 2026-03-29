package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
)

type MarketLister interface {
	List(ctx context.Context) ([]domain.Market, error)
}

type OrderPlacer interface {
	PlaceMarketBuy(ctx context.Context, input orderservice.PlaceMarketBuyInput) (domain.Order, error)
}

type PortfolioGetter interface {
	GetSummary(ctx context.Context, walletID string) (domain.PortfolioSummary, error)
}

func NewRouter(markets MarketLister, orders OrderPlacer, portfolios PortfolioGetter) http.Handler {
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

		writeJSON(w, http.StatusOK, map[string]any{
			"markets": items,
		})
	})

	mux.HandleFunc("GET /api/v1/portfolios/", func(w http.ResponseWriter, r *http.Request) {
		walletID := strings.TrimPrefix(r.URL.Path, "/api/v1/portfolios/")
		if walletID == "" {
			writeError(w, http.StatusBadRequest, "wallet ID is required")
			return
		}

		summary, err := portfolios.GetSummary(r.Context(), walletID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to load portfolio")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"portfolio": summary,
		})
	})

	mux.HandleFunc("POST /api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			UserID        string  `json:"user_id"`
			WalletID      string  `json:"wallet_id"`
			MarketSymbol  string  `json:"market_symbol"`
			QuoteAmount   float64 `json:"quote_amount"`
			ExpectedPrice float64 `json:"expected_price"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON payload")
			return
		}

		order, err := orders.PlaceMarketBuy(r.Context(), orderservice.PlaceMarketBuyInput{
			UserID:        payload.UserID,
			WalletID:      payload.WalletID,
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

		writeJSON(w, http.StatusCreated, map[string]any{
			"order": order,
		})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
