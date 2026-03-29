package main

import (
	"log"
	"net/http"

	"github.com/aidun/tradelab/backend/internal/config"
	httpapi "github.com/aidun/tradelab/backend/internal/http"
	marketservice "github.com/aidun/tradelab/backend/internal/service/market"
	orderservice "github.com/aidun/tradelab/backend/internal/service/order"
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
	orderRepository := postgres.NewOrderRepository(db)

	marketService := marketservice.NewService(marketRepository)
	orderService := orderservice.NewService(marketRepository, balanceRepository, orderRepository)

	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: httpapi.NewRouter(marketService, orderService),
	}

	log.Printf("TradeLab backend listening on %s", cfg.HTTPAddress)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start server: %v", err)
	}
}
