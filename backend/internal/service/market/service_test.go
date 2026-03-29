package market

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aidun/tradelab/backend/internal/domain"
)

func TestListCandlesFetchesRemoteMarketData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/uiKlines" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("symbol"); got != "XRPUSDT" {
			t.Fatalf("expected symbol XRPUSDT, got %s", got)
		}

		if got := r.URL.Query().Get("interval"); got != "1h" {
			t.Fatalf("expected interval 1h, got %s", got)
		}

		if got := r.URL.Query().Get("limit"); got != "48" {
			t.Fatalf("expected limit 48, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			[1743242400000, "0.6200", "0.6400", "0.6150", "0.6320", "1250000.0", 1743245999999, "789000.0", 8342, "0", "0", "0"],
			[1743246000000, "0.6320", "0.6450", "0.6280", "0.6410", "1430000.0", 1743249599999, "912000.0", 9021, "0", "0", "0"]
		]`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{
		market: domain.Market{
			ID:         "market-1",
			Symbol:     "XRP/USDT",
			BaseAsset:  "XRP",
			QuoteAsset: "USDT",
		},
	}, server.URL)
	service.client = server.Client()

	candles, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(candles) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(candles))
	}

	if candles[1].ClosePrice != 0.641 {
		t.Fatalf("expected close price 0.641, got %f", candles[1].ClosePrice)
	}
}

func TestGetSpotPriceFetchesRemoteMarketPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/ticker/price" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		if got := r.URL.Query().Get("symbol"); got != "XRPUSDT" {
			t.Fatalf("expected symbol XRPUSDT, got %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"XRPUSDT","price":"0.6543"}`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{
		market: domain.Market{
			ID:         "market-1",
			Symbol:     "XRP/USDT",
			BaseAsset:  "XRP",
			QuoteAsset: "USDT",
		},
	}, server.URL)
	service.client = server.Client()

	price, err := service.GetSpotPrice(context.Background(), "XRP/USDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if price != 0.6543 {
		t.Fatalf("expected price 0.6543, got %f", price)
	}
}

type fakeMarketRepository struct {
	market  domain.Market
	markets []domain.Market
	err     error
}

func (f fakeMarketRepository) List(context.Context) ([]domain.Market, error) {
	return f.markets, f.err
}

func (f fakeMarketRepository) GetBySymbol(context.Context, string) (domain.Market, error) {
	return f.market, f.err
}
