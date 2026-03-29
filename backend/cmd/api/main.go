package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/aidun/tradelab/backend/internal/config"
	httpapi "github.com/aidun/tradelab/backend/internal/http"
	"github.com/aidun/tradelab/backend/internal/logging"
	historyservice "github.com/aidun/tradelab/backend/internal/service/history"
	marketservice "github.com/aidun/tradelab/backend/internal/service/market"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	portfolioservice "github.com/aidun/tradelab/backend/internal/service/portfolio"
	sessionservice "github.com/aidun/tradelab/backend/internal/service/session"
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
	sessionRepository := postgres.NewDemoSessionRepository(db)

	marketService := marketservice.NewService(marketRepository, cfg.MarketDataBaseURL, logging.NewJSONLogger("market_service"))
	orderService := orderservice.NewService(marketRepository, balanceRepository, portfolioRepository, marketService, logging.NewJSONLogger("order_service"))
	portfolioService := portfolioservice.NewService(portfolioRepository, logging.NewJSONLogger("portfolio_service"))
	historyService := historyservice.NewService(portfolioRepository, logging.NewJSONLogger("history_service"))
	sessionService := sessionservice.NewService(sessionRepository, logging.NewJSONLogger("session_service"))
	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: httpapi.NewRouter(marketService, marketService, orderService, portfolioService, historyService, historyService, sessionService, logging.NewJSONLogger("http_api")),
	}

	logger.Info("backend listening", "operation", "startup.listen", "address", cfg.HTTPAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server terminated unexpectedly", "operation", "startup.serve", "error", err)
		os.Exit(1)
	}
}
