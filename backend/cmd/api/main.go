package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aidun/tradelab/backend/internal/config"
	httpapi "github.com/aidun/tradelab/backend/internal/http"
	"github.com/aidun/tradelab/backend/internal/logging"
	accountservice "github.com/aidun/tradelab/backend/internal/service/account"
	backtestservice "github.com/aidun/tradelab/backend/internal/service/backtest"
	historyservice "github.com/aidun/tradelab/backend/internal/service/history"
	marketservice "github.com/aidun/tradelab/backend/internal/service/market"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	portfolioservice "github.com/aidun/tradelab/backend/internal/service/portfolio"
	sessionservice "github.com/aidun/tradelab/backend/internal/service/session"
	strategyservice "github.com/aidun/tradelab/backend/internal/service/strategy"
	"github.com/aidun/tradelab/backend/internal/store/postgres"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With("service", "tradelab-api")

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open postgres", "operation", "startup.open_postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	marketRepository := postgres.NewMarketRepository(db)
	balanceRepository := postgres.NewBalanceRepository(db)
	portfolioRepository := postgres.NewPortfolioRepository(db)
	strategyRepository := postgres.NewStrategyRepository(db)
	sessionRepository := postgres.NewDemoSessionRepository(db)
	appSessionRepository := postgres.NewAppSessionRepository(db)
	registeredAccountRepository := postgres.NewRegisteredAccountRepository(db)

	clerkVerifier, err := accountservice.NewClerkTokenVerifier(context.Background(), accountservice.ClerkVerifierConfig{
		JWKSURL:  cfg.ClerkJWKSURL,
		Issuer:   cfg.ClerkIssuerURL,
		MockMode: cfg.AuthMockMode,
	})
	if err != nil {
		logger.Error("failed to initialize clerk verifier", "operation", "startup.clerk_verifier", "error", err)
		os.Exit(1)
	}

	marketService := marketservice.NewService(marketRepository, cfg.MarketDataBaseURL, logging.NewJSONLogger("market_service"))
	orderService := orderservice.NewService(marketRepository, balanceRepository, portfolioRepository, marketService, logging.NewJSONLogger("order_service"))
	portfolioService := portfolioservice.NewService(portfolioRepository, marketService, logging.NewJSONLogger("portfolio_service"))
	historyService := historyservice.NewService(portfolioRepository, logging.NewJSONLogger("history_service"))
	backtestService := backtestservice.NewService(marketRepository, marketService)
	sessionService := sessionservice.NewService(sessionRepository, appSessionRepository, logging.NewJSONLogger("session_service"))
	accountService := accountservice.NewService(registeredAccountRepository, clerkVerifier, logging.NewJSONLogger("account_service"))
	strategyService := strategyservice.NewService(marketRepository, strategyRepository, marketService, portfolioService, orderService, logging.NewJSONLogger("strategy_service"))
	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: httpapi.NewRouter(marketService, marketService, backtestService, orderService, portfolioService, historyService, historyService, strategyService, sessionService, accountService, logging.NewJSONLogger("http_api")),
	}

	if cfg.StrategyEngineEnabled {
		go runStrategyLoop(context.Background(), strategyService, cfg.StrategyEngineTick, logger)
	}

	logger.Info("backend listening", "operation", "startup.listen", "address", cfg.HTTPAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server terminated unexpectedly", "operation", "startup.serve", "error", err)
		os.Exit(1)
	}
}

func runStrategyLoop(ctx context.Context, strategies interface {
	RunOnce(context.Context, int) error
}, tick time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	logger.Info("strategy engine started", "operation", "strategy_engine.start", "tick", tick.String())

	for {
		if err := strategies.RunOnce(ctx, 25); err != nil {
			logger.Error("strategy engine tick failed", "operation", "strategy_engine.tick", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
