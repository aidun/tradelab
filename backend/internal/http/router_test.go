package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
)

func TestHealthRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeOrderPlacer{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListMarketsRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/markets", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{
		markets: []domain.Market{
			{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT", Exchange: "demo"},
		},
	}, fakeOrderPlacer{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if body := recorder.Body.String(); !strings.Contains(body, "XRP/USDT") {
		t.Fatalf("expected market symbol in response body, got %s", body)
	}
}

func TestCreateOrderRoute(t *testing.T) {
	body := bytes.NewBufferString(`{"user_id":"user-1","wallet_id":"wallet-1","market_symbol":"XRP/USDT","quote_amount":50,"expected_price":0.67}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", body)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeOrderPlacer{
		order: domain.Order{
			ID:           "order-1",
			UserID:       "user-1",
			WalletID:     "wallet-1",
			MarketID:     "market-1",
			MarketSymbol: "XRP/USDT",
			QuoteAmount:  50,
			Status:       domain.OrderStatusPending,
		},
	}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	if body := recorder.Body.String(); !strings.Contains(body, "order-1") {
		t.Fatalf("expected order ID in response body, got %s", body)
	}
}

type fakeMarketLister struct {
	markets []domain.Market
	err     error
}

func (f fakeMarketLister) List(context.Context) ([]domain.Market, error) {
	return f.markets, f.err
}

type fakeOrderPlacer struct {
	order domain.Order
	err   error
}

func (f fakeOrderPlacer) PlaceMarketBuy(context.Context, orderservice.PlaceMarketBuyInput) (domain.Order, error) {
	return f.order, f.err
}
