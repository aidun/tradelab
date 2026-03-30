package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	strategyservice "github.com/aidun/tradelab/backend/internal/service/strategy"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestHealthRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestCreateDemoSessionRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/demo", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{
			ID:        "session-1",
			UserID:    "user-1",
			WalletID:  "wallet-1",
			Token:     "token-1",
			ExpiresAt: time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
		},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"token":"token-1"`) {
		t.Fatalf("expected token in response, got %s", recorder.Body.String())
	}
}

func TestListMarketsRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/markets", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{
		markets: []domain.Market{
			{ID: "market-1", Symbol: "XRP/USDT", BaseAsset: "XRP", QuoteAsset: "USDT", Exchange: "demo"},
		},
	}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestPortfolioRouteRequiresSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/wallet-1", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestPortfolioRouteRejectsForeignWallet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/wallet-2", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestPortfolioRouteReturnsOwnedWallet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/wallet-1", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{
		summary: domain.PortfolioSummary{WalletID: "wallet-1", BaseCurrency: "USDT", TotalValue: 10000},
	}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListOrdersRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{
		orders: []domain.Order{{ID: "order-1", WalletID: "wallet-1", MarketSymbol: "XRP/USDT"}},
	}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "order-1") {
		t.Fatalf("expected order payload in response, got %s", recorder.Body.String())
	}
}

func TestListActivityRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{
		activity: []domain.ActivityLog{{ID: "log-1", WalletID: "wallet-1", Title: "Demo buy recorded", CreatedAt: time.Now()}},
	}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestListActivityRouteReturnsEmptyArrayWhenNoEntries(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"activity":[]`) {
		t.Fatalf("expected empty activity array in response, got %s", recorder.Body.String())
	}
}

func TestListStrategiesRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategies?market_symbol=XRP/USDT", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{
		strategies: []domain.Strategy{{ID: "strategy-1", WalletID: "wallet-1", MarketSymbol: "XRP/USDT", Status: domain.StrategyStatusActive}},
	}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "strategy-1") {
		t.Fatalf("expected strategy payload in response, got %s", recorder.Body.String())
	}
}

func TestListStrategiesRouteReturnsEmptyArrayWhenNoEntries(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/strategies", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"strategies":[]`) {
		t.Fatalf("expected empty strategies array in response, got %s", recorder.Body.String())
	}
}

func TestCreateStrategyRoute(t *testing.T) {
	body := bytes.NewBufferString(`{"market_symbol":"XRP/USDT","status":"active","config":{"dipBuy":{"enabled":true,"dipPercent":5,"spendQuoteAmount":100},"takeProfit":{"enabled":true,"triggerPercent":8},"stopLoss":{"enabled":true,"triggerPercent":3}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/strategies", body)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{
		strategy: domain.Strategy{ID: "strategy-1", WalletID: "wallet-1", MarketSymbol: "XRP/USDT", Status: domain.StrategyStatusActive},
	}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

func TestListMarketCandlesRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/markets/XRP%2FUSDT/candles?interval=1h&limit=2", nil)
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{
		feed: domain.CandleFeed{
			Candles: []domain.Candle{
				{OpenPrice: 0.62, HighPrice: 0.64, LowPrice: 0.61, ClosePrice: 0.63},
				{OpenPrice: 0.63, HighPrice: 0.65, LowPrice: 0.62, ClosePrice: 0.64},
			},
			Meta: domain.MarketDataMeta{Source: "stale", GeneratedAt: time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)},
		},
	}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "\"candles\"") {
		t.Fatalf("expected candles payload in response, got %s", recorder.Body.String())
	}

	if !strings.Contains(recorder.Body.String(), "\"source\":\"stale\"") {
		t.Fatalf("expected stale candle metadata in response, got %s", recorder.Body.String())
	}
}

func TestCreateOrderRoute(t *testing.T) {
	body := bytes.NewBufferString(`{"market_symbol":"XRP/USDT","side":"buy","quote_amount":50}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", body)
	req.Header.Set("Authorization", "Bearer token-1")
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
	}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		session: domain.DemoSession{UserID: "user-1", WalletID: "wallet-1"},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

func TestOrdersRouteReturnsInternalErrorForUnexpectedSessionFailure(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.Header.Set("Authorization", "Bearer token-1")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		err: errors.New("store unavailable"),
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestBootstrapRegisteredAccountRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/bootstrap", nil)
	req.Header.Set("Authorization", "Bearer registered-token")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		appSession: domain.AppSession{
			ID:            "app-session-1",
			Token:         "cookie-token",
			IdleExpiresAt: time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
			LastUsedAt:    time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
			PrincipalKind: domain.PrincipalKindRegistered,
			UserID:        "user-registered",
			WalletID:      "wallet-registered",
		},
	}, fakeRegisteredAccountManager{
		account: domain.RegisteredAccount{
			UserID:      "user-registered",
			WalletID:    "wallet-registered",
			ClerkUserID: "clerk-user-1",
			Email:       "trader@example.com",
			DisplayName: "Trader Example",
		},
	}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"mode":"registered"`) {
		t.Fatalf("expected registered mode in response, got %s", recorder.Body.String())
	}

	if !strings.Contains(recorder.Header().Get("Set-Cookie"), registeredSessionCookieName+"=cookie-token") {
		t.Fatalf("expected registered session cookie to be set, got %s", recorder.Header().Get("Set-Cookie"))
	}
}

func TestUpgradeGuestAccountRoute(t *testing.T) {
	body := bytes.NewBufferString(`{"preserve_guest_data":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/upgrade", body)
	req.Header.Set("Authorization", "Bearer registered-token")
	req.Header.Set("X-TradeLab-Guest-Token", "guest-token")
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		appSession: domain.AppSession{
			ID:            "app-session-1",
			Token:         "cookie-token",
			IdleExpiresAt: time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC),
			LastUsedAt:    time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC),
			PrincipalKind: domain.PrincipalKindRegistered,
			UserID:        "user-registered",
			WalletID:      "wallet-registered",
		},
	}, fakeRegisteredAccountManager{
		account: domain.RegisteredAccount{
			UserID:      "user-registered",
			WalletID:    "wallet-registered",
			ClerkUserID: "clerk-user-1",
			Email:       "trader@example.com",
			DisplayName: "Trader Example",
		},
	}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestOrdersRouteSupportsRegisteredAccounts(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	req.AddCookie(&http.Cookie{Name: registeredSessionCookieName, Value: "cookie-token"})
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{
		orders: []domain.Order{{ID: "order-registered", WalletID: "wallet-registered", MarketSymbol: "BTC/USDT"}},
	}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{
		appSession: domain.AppSession{
			ID:                "app-session-1",
			UserID:            "user-registered",
			WalletID:          "wallet-registered",
			PrincipalKind:     domain.PrincipalKindRegistered,
			IdleExpiresAt:     time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
			AbsoluteExpiresAt: time.Date(2026, 4, 29, 11, 0, 0, 0, time.UTC),
		},
	}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestLogoutRouteRevokesRegisteredCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/logout", nil)
	req.AddCookie(&http.Cookie{Name: registeredSessionCookieName, Value: "cookie-token"})
	recorder := httptest.NewRecorder()

	NewRouter(fakeMarketLister{}, fakeMarketLister{}, fakeOrderPlacer{}, fakePortfolioGetter{}, fakeOrderHistoryLister{}, fakeActivityHistoryLister{}, fakeStrategyManager{}, fakeSessionManager{}, fakeRegisteredAccountManager{}, discardLogger()).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Header().Get("Set-Cookie"), registeredSessionCookieName+"=") {
		t.Fatalf("expected registered session cookie to be cleared, got %s", recorder.Header().Get("Set-Cookie"))
	}
}

type fakeMarketLister struct {
	markets []domain.Market
	feed    domain.CandleFeed
	err     error
}

func (f fakeMarketLister) List(context.Context) ([]domain.Market, error) { return f.markets, f.err }
func (f fakeMarketLister) ListCandles(context.Context, string, string, int) (domain.CandleFeed, error) {
	return f.feed, f.err
}

type fakeOrderPlacer struct {
	order domain.Order
	err   error
}

func (f fakeOrderPlacer) PlaceMarketOrder(context.Context, orderservice.PlaceMarketOrderInput) (domain.Order, error) {
	return f.order, f.err
}

type fakePortfolioGetter struct {
	summary domain.PortfolioSummary
	err     error
}

func (f fakePortfolioGetter) GetSummary(context.Context, string, domain.AccountingMode) (domain.PortfolioSummary, error) {
	return f.summary, f.err
}

type fakeOrderHistoryLister struct {
	orders []domain.Order
	err    error
}

func (f fakeOrderHistoryLister) ListOrders(context.Context, string, int, string, domain.AccountingMode) ([]domain.Order, error) {
	return f.orders, f.err
}

type fakeActivityHistoryLister struct {
	activity []domain.ActivityLog
	err      error
}

func (f fakeActivityHistoryLister) ListActivity(context.Context, string, int, string) ([]domain.ActivityLog, error) {
	return f.activity, f.err
}

type fakeStrategyManager struct {
	strategies []domain.Strategy
	strategy   domain.Strategy
	err        error
}

func (f fakeStrategyManager) ListStrategies(context.Context, string, string) ([]domain.Strategy, error) {
	return f.strategies, f.err
}

func (f fakeStrategyManager) UpsertStrategy(context.Context, strategyservice.UpsertStrategyInput) (domain.Strategy, error) {
	return f.strategy, f.err
}

func (f fakeStrategyManager) PatchStrategy(context.Context, strategyservice.PatchStrategyInput) (domain.Strategy, error) {
	return f.strategy, f.err
}

type fakeSessionManager struct {
	session      domain.DemoSession
	appSession   domain.AppSession
	err          error
	appErr       error
	revokeErr    error
	revokedToken string
}

func (f fakeSessionManager) CreateDemoSession(context.Context) (domain.DemoSession, error) {
	return f.session, f.err
}

func (f fakeSessionManager) Authenticate(context.Context, string) (domain.DemoSession, error) {
	return f.session, f.err
}

func (f fakeSessionManager) CreateRegisteredSession(context.Context, string, string) (domain.AppSession, error) {
	return f.appSession, f.appErr
}

func (f fakeSessionManager) AuthenticateRegistered(context.Context, string) (domain.AppSession, error) {
	return f.appSession, f.appErr
}

func (f fakeSessionManager) RevokeRegisteredSession(_ context.Context, token string) error {
	f.revokedToken = token
	return f.revokeErr
}

type fakeRegisteredAccountManager struct {
	account domain.RegisteredAccount
	err     error
}

func (f fakeRegisteredAccountManager) BootstrapRegisteredAccount(context.Context, string) (domain.RegisteredAccount, error) {
	return f.account, f.err
}

func (f fakeRegisteredAccountManager) UpgradeGuestSession(context.Context, string, string, bool) (domain.RegisteredAccount, error) {
	return f.account, f.err
}
