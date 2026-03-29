package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
)

func TestHealthRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

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
	}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestPortfolioRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/wallet-1", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{
		summary: domain.PortfolioSummary{WalletID: "wallet-1", BaseCurrency: "USDT", TotalValue: 10000},
	}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListOrdersRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?wallet_id=wallet-1", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{
		orders: []domain.Order{{ID: "order-1", WalletID: "wallet-1", MarketSymbol: "XRP/USDT"}},
	}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "order-1") {
		t.Fatalf("expected order payload in response, got %s", recorder.Body.String())
	}
}

func TestListActivityRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?wallet_id=wallet-1", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{
		activity: []domain.ActivityLog{{ID: "log-1", WalletID: "wallet-1", Title: "Demo buy recorded", CreatedAt: time.Now()}},
	}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListMarketCandlesRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/markets/XRP%2FUSDT/candles?interval=1h&limit=2", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{
		candles: []domain.Candle{
			{OpenPrice: 0.62, HighPrice: 0.64, LowPrice: 0.61, ClosePrice: 0.63},
			{OpenPrice: 0.63, HighPrice: 0.65, LowPrice: 0.62, ClosePrice: 0.64},
		},
	}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "\"candles\"") {
		t.Fatalf("expected candles payload in response, got %s", recorder.Body.String())
	}
}

func TestCreateOrderRoute(t *testing.T) {
	body := bytes.NewBufferString(`{"user_id":"user-1","wallet_id":"wallet-1","market_symbol":"XRP/USDT","quote_amount":50,"expected_price":0.67}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", body)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{
		order: domain.Order{
			ID:           "order-1",
			UserID:       "user-1",
			WalletID:     "wallet-1",
			MarketID:     "market-1",
			MarketSymbol: "XRP/USDT",
			QuoteAmount:  50,
			Status:       domain.OrderStatusFilled,
		},
	}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

type fakeMarketLister struct {
	markets []domain.Market
	candles []domain.Candle
	err     error
}

func (f fakeMarketLister) List(context.Context) ([]domain.Market, error) { return f.markets, f.err }
func (f fakeMarketLister) ListCandles(context.Context, string, string, int) ([]domain.Candle, error) {
	return f.candles, f.err
}

type fakeOrderPlacer struct {
	order domain.Order
	err   error
}

func (f fakeOrderPlacer) PlaceMarketBuy(context.Context, orderservice.PlaceMarketBuyInput) (domain.Order, error) {
	return f.order, f.err
}

type fakePortfolioGetter struct {
	summary domain.PortfolioSummary
	err     error
}

func (f fakePortfolioGetter) GetSummary(context.Context, string) (domain.PortfolioSummary, error) {
	return f.summary, f.err
}

type fakeOrderHistoryLister struct {
	orders []domain.Order
	err    error
}

func (f fakeOrderHistoryLister) ListOrders(context.Context, string, int) ([]domain.Order, error) {
	return f.orders, f.err
}

type fakeActivityHistoryLister struct {
	activity []domain.ActivityLog
	err      error
}

func (f fakeActivityHistoryLister) ListActivity(context.Context, string, int) ([]domain.ActivityLog, error) {
	return f.activity, f.err
}
