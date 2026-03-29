package main

import (
	"log"
	"net/http"

	"github.com/aidun/tradelab/backend/internal/config"
	httpapi "github.com/aidun/tradelab/backend/internal/http"
	historyservice "github.com/aidun/tradelab/backend/internal/service/history"
	marketservice "github.com/aidun/tradelab/backend/internal/service/market"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
	portfolioservice "github.com/aidun/tradelab/backend/internal/service/portfolio"
	"github.com/aidun/tradelab/backend/internal/store/postgres"
)

func main() {
	cfg := config.Load()

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	marketRepository := postgres.NewMarketRepository(db)
	balanceRepository := postgres.NewBalanceRepository(db)
	portfolioRepository := postgres.NewPortfolioRepository(db)

	marketService := marketservice.NewService(marketRepository, cfg.MarketDataBaseURL)
	orderService := orderservice.NewService(marketRepository, balanceRepository, portfolioRepository)
	portfolioService := portfolioservice.NewService(portfolioRepository)
	historyService := historyservice.NewService(portfolioRepository)
	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: httpapi.NewRouter(marketService, marketService, orderService, portfolioService, historyService, historyService),
	}

	log.Printf("TradeLab backend listening on %s", cfg.HTTPAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start server: %v", err)
	}
}
