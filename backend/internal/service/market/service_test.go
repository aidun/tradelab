package market

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

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
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	feed, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(feed.Candles) != 2 {
		t.Fatalf("expected 2 candles, got %d", len(feed.Candles))
	}

	if feed.Candles[1].ClosePrice != 0.641 {
		t.Fatalf("expected close price 0.641, got %f", feed.Candles[1].ClosePrice)
	}

	if feed.Meta.Source != "fresh" {
		t.Fatalf("expected fresh source, got %s", feed.Meta.Source)
	}
}

func TestListCandlesUsesFreshCache(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			[1743242400000, "0.6200", "0.6400", "0.6150", "0.6320", "1250000.0", 1743245999999, "789000.0", 8342, "0", "0", "0"]
		]`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	clock := &stepClock{times: []time.Time{
		time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 5, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 10, 0, time.UTC),
	}}
	service.clock = clock

	first, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	second, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected one upstream call, got %d", calls)
	}

	if second.Meta.Source != "fresh" {
		t.Fatalf("expected fresh source from cache, got %s", second.Meta.Source)
	}

	if first.Meta.GeneratedAt != second.Meta.GeneratedAt {
		t.Fatalf("expected cached generated time to match")
	}
}

func TestListCandlesFallsBackToStaleCacheOnUpstreamFailure(t *testing.T) {
	var shouldFail atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail.Load() {
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			[1743242400000, "0.6200", "0.6400", "0.6150", "0.6320", "1250000.0", 1743245999999, "789000.0", 8342, "0", "0", "0"]
		]`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	clock := &stepClock{times: []time.Time{
		time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 1, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 20, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 40, 0, time.UTC),
	}}
	service.clock = clock

	if _, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48); err != nil {
		t.Fatalf("expected initial fetch to succeed, got %v", err)
	}

	shouldFail.Store(true)

	feed, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err != nil {
		t.Fatalf("expected stale fallback, got %v", err)
	}

	if feed.Meta.Source != "stale" {
		t.Fatalf("expected stale source, got %s", feed.Meta.Source)
	}
}

func TestListCandlesReturnsErrorWhenNoFallbackIsAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	_, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestListCandlesReturnsErrorForMalformedPayloadWithoutFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"invalid":true}`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	_, err := service.ListCandles(context.Background(), "XRP/USDT", "1h", 48)
	if err == nil {
		t.Fatal("expected an error, got nil")
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

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	service.clock = fakeClock{now: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)}

	price, err := service.GetSpotPrice(context.Background(), "XRP/USDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if price != 0.6543 {
		t.Fatalf("expected price 0.6543, got %f", price)
	}
}

func TestGetSpotPriceUsesFreshCacheAndStaleFallback(t *testing.T) {
	var calls int32
	var shouldFail atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if shouldFail.Load() {
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"XRPUSDT","price":"0.6543"}`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	clock := &stepClock{times: []time.Time{
		time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 1, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 7, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 20, 0, time.UTC),
	}}
	service.clock = clock

	first, err := service.GetSpotPrice(context.Background(), "XRP/USDT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	second, err := service.GetSpotPrice(context.Background(), "XRP/USDT")
	if err != nil {
		t.Fatalf("expected cached price, got %v", err)
	}

	shouldFail.Store(true)

	third, err := service.GetSpotPrice(context.Background(), "XRP/USDT")
	if err != nil {
		t.Fatalf("expected stale fallback, got %v", err)
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected two upstream calls, got %d", calls)
	}

	if first != second || second != third {
		t.Fatalf("expected cached values to match")
	}
}

func TestGetSpotPriceReturnsErrorAfterStaleWindowExpires(t *testing.T) {
	var shouldFail atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldFail.Load() {
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"XRPUSDT","price":"0.6543"}`))
	}))
	defer server.Close()

	service := NewService(fakeMarketRepository{market: demoMarket()}, server.URL)
	service.client = server.Client()
	clock := &stepClock{times: []time.Time{
		time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 29, 12, 0, 40, 0, time.UTC),
	}}
	service.clock = clock

	if _, err := service.GetSpotPrice(context.Background(), "XRP/USDT"); err != nil {
		t.Fatalf("expected initial fetch to succeed, got %v", err)
	}

	shouldFail.Store(true)

	if _, err := service.GetSpotPrice(context.Background(), "XRP/USDT"); err == nil {
		t.Fatal("expected an error after stale window expiry, got nil")
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

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type stepClock struct {
	times []time.Time
	index int
}

func (c *stepClock) Now() time.Time {
	if len(c.times) == 0 {
		return time.Time{}
	}

	if c.index >= len(c.times) {
		return c.times[len(c.times)-1]
	}

	value := c.times[c.index]
	c.index++
	return value
}

func demoMarket() domain.Market {
	return domain.Market{
		ID:          "market-1",
		Symbol:      "XRP/USDT",
		BaseAsset:   "XRP",
		QuoteAsset:  "USDT",
		MinNotional: 10,
	}
}
