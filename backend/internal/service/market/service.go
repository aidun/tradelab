package market

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	markets           store.MarketRepository
	marketDataBaseURL string
	client            *http.Client
	logger            *slog.Logger
	clock             Clock
	cacheMu           sync.RWMutex
	spotCache         map[string]spotPriceCacheEntry
	candleCache       map[string]candleCacheEntry
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type spotPriceCacheEntry struct {
	price       float64
	generatedAt time.Time
	freshUntil  time.Time
	staleUntil  time.Time
}

type candleCacheEntry struct {
	candles     []domain.Candle
	generatedAt time.Time
	freshUntil  time.Time
	staleUntil  time.Time
}

const (
	spotPriceFreshTTL = 5 * time.Second
	spotPriceStaleTTL = 30 * time.Second
	candleFreshTTL    = 15 * time.Second
	candleStaleTTL    = 2 * time.Minute
)

func NewService(markets store.MarketRepository, marketDataBaseURL string, logger *slog.Logger) *Service {
	return &Service{
		markets:           markets,
		marketDataBaseURL: strings.TrimRight(marketDataBaseURL, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:      logger,
		clock:       realClock{},
		spotCache:   map[string]spotPriceCacheEntry{},
		candleCache: map[string]candleCacheEntry{},
	}
}

func (s *Service) List(ctx context.Context) ([]domain.Market, error) {
	return s.markets.List(ctx)
}

func (s *Service) GetSpotPrice(ctx context.Context, marketSymbol string) (float64, error) {
	market, err := s.markets.GetBySymbol(ctx, marketSymbol)
	if err != nil {
		s.logError("get_spot_price.market_lookup_failed", err, "market_symbol", marketSymbol)
		return 0, fmt.Errorf("get market: %w", err)
	}

	now := s.clock.Now()
	if entry, ok := s.getSpotCacheEntry(market.Symbol); ok && now.Before(entry.freshUntil) {
		s.logInfo("get_spot_price.cache_hit", "market_symbol", market.Symbol, "source", "fresh")
		return entry.price, nil
	}

	price, generatedAt, err := s.fetchSpotPrice(ctx, market.Symbol)
	if err == nil {
		s.storeSpotCacheEntry(market.Symbol, price, generatedAt)
		s.logInfo("get_spot_price.upstream_success", "market_symbol", market.Symbol, "source", "fresh", "generated_at", generatedAt)
		return price, nil
	}

	if entry, ok := s.getSpotCacheEntry(market.Symbol); ok && now.Before(entry.staleUntil) {
		s.logInfo("get_spot_price.cache_fallback", "market_symbol", market.Symbol, "source", "stale", "generated_at", entry.generatedAt, "error", err)
		return entry.price, nil
	}

	s.logError("get_spot_price.unavailable", err, "market_symbol", market.Symbol)
	return 0, err
}

func (s *Service) ListCandles(ctx context.Context, marketSymbol string, interval string, limit int) (domain.CandleFeed, error) {
	if interval == "" {
		interval = "1h"
	}

	if limit <= 0 || limit > 200 {
		limit = 48
	}

	market, err := s.markets.GetBySymbol(ctx, marketSymbol)
	if err != nil {
		s.logError("list_candles.market_lookup_failed", err, "market_symbol", marketSymbol)
		return domain.CandleFeed{}, fmt.Errorf("get market: %w", err)
	}

	cacheKey := candleCacheKey(market.Symbol, interval, limit)
	now := s.clock.Now()
	if entry, ok := s.getCandleCacheEntry(cacheKey); ok && now.Before(entry.freshUntil) {
		s.logInfo("list_candles.cache_hit", "market_symbol", market.Symbol, "interval", interval, "limit", limit, "source", "fresh")
		return domain.CandleFeed{
			Candles: cloneCandles(entry.candles),
			Meta: domain.MarketDataMeta{
				Source:      "fresh",
				GeneratedAt: entry.generatedAt,
			},
		}, nil
	}

	candles, generatedAt, err := s.fetchCandles(ctx, market.Symbol, interval, limit)
	if err == nil {
		s.storeCandleCacheEntry(cacheKey, candles, generatedAt)
		s.logInfo("list_candles.upstream_success", "market_symbol", market.Symbol, "interval", interval, "limit", limit, "source", "fresh", "generated_at", generatedAt)
		return domain.CandleFeed{
			Candles: candles,
			Meta: domain.MarketDataMeta{
				Source:      "fresh",
				GeneratedAt: generatedAt,
			},
		}, nil
	}

	if entry, ok := s.getCandleCacheEntry(cacheKey); ok && now.Before(entry.staleUntil) {
		s.logInfo("list_candles.cache_fallback", "market_symbol", market.Symbol, "interval", interval, "limit", limit, "source", "stale", "generated_at", entry.generatedAt, "error", err)
		return domain.CandleFeed{
			Candles: cloneCandles(entry.candles),
			Meta: domain.MarketDataMeta{
				Source:      "stale",
				GeneratedAt: entry.generatedAt,
			},
		}, nil
	}

	s.logError("list_candles.unavailable", err, "market_symbol", market.Symbol, "interval", interval, "limit", limit)
	return domain.CandleFeed{}, err
}

func (s *Service) fetchSpotPrice(ctx context.Context, marketSymbol string) (float64, time.Time, error) {
	query := url.Values{}
	query.Set("symbol", strings.ReplaceAll(marketSymbol, "/", ""))

	endpoint := fmt.Sprintf("%s/api/v3/ticker/price?%s", s.marketDataBaseURL, query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("create market price request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("request market price: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return 0, time.Time{}, fmt.Errorf("market price request failed with status %d", response.StatusCode)
	}

	var payload struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return 0, time.Time{}, fmt.Errorf("decode market price: %w", err)
	}

	price, err := strconv.ParseFloat(payload.Price, 64)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("parse market price: %w", err)
	}

	return price, s.clock.Now(), nil
}

func (s *Service) fetchCandles(ctx context.Context, marketSymbol string, interval string, limit int) ([]domain.Candle, time.Time, error) {
	query := url.Values{}
	query.Set("symbol", strings.ReplaceAll(marketSymbol, "/", ""))
	query.Set("interval", interval)
	query.Set("limit", strconv.Itoa(limit))

	endpoint := fmt.Sprintf("%s/api/v3/uiKlines?%s", s.marketDataBaseURL, query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("create market data request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("request market data: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, time.Time{}, fmt.Errorf("market data request failed with status %d", response.StatusCode)
	}

	var payload [][]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, time.Time{}, fmt.Errorf("decode market data: %w", err)
	}

	candles := make([]domain.Candle, 0, len(payload))
	for _, row := range payload {
		candle, err := parseCandle(row)
		if err != nil {
			return nil, time.Time{}, err
		}
		candles = append(candles, candle)
	}

	return candles, s.clock.Now(), nil
}

func candleCacheKey(symbol string, interval string, limit int) string {
	return fmt.Sprintf("%s|%s|%d", symbol, interval, limit)
}

func cloneCandles(candles []domain.Candle) []domain.Candle {
	items := make([]domain.Candle, len(candles))
	copy(items, candles)
	return items
}

func (s *Service) getSpotCacheEntry(key string) (spotPriceCacheEntry, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	entry, ok := s.spotCache[key]
	return entry, ok
}

func (s *Service) storeSpotCacheEntry(key string, price float64, generatedAt time.Time) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.spotCache[key] = spotPriceCacheEntry{
		price:       price,
		generatedAt: generatedAt,
		freshUntil:  generatedAt.Add(spotPriceFreshTTL),
		staleUntil:  generatedAt.Add(spotPriceStaleTTL),
	}
}

func (s *Service) getCandleCacheEntry(key string) (candleCacheEntry, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	entry, ok := s.candleCache[key]
	return entry, ok
}

func (s *Service) storeCandleCacheEntry(key string, candles []domain.Candle, generatedAt time.Time) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Cache a clone so callers cannot mutate the stored fallback copy through a shared slice.
	s.candleCache[key] = candleCacheEntry{
		candles:     cloneCandles(candles),
		generatedAt: generatedAt,
		freshUntil:  generatedAt.Add(candleFreshTTL),
		staleUntil:  generatedAt.Add(candleStaleTTL),
	}
}

func (s *Service) logInfo(operation string, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Info(operation, append([]any{"operation", operation}, args...)...)
}

func (s *Service) logError(operation string, err error, args ...any) {
	if s.logger == nil {
		return
	}

	s.logger.Error(operation, append([]any{"operation", operation, "error", err}, args...)...)
}

func parseCandle(row []any) (domain.Candle, error) {
	if len(row) < 9 {
		return domain.Candle{}, fmt.Errorf("unexpected candle payload length: %d", len(row))
	}

	openTime, err := parseUnixMilliseconds(row[0])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse open time: %w", err)
	}

	closeTime, err := parseUnixMilliseconds(row[6])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse close time: %w", err)
	}

	openPrice, err := parseStringFloat(row[1])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse open price: %w", err)
	}

	highPrice, err := parseStringFloat(row[2])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse high price: %w", err)
	}

	lowPrice, err := parseStringFloat(row[3])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse low price: %w", err)
	}

	closePrice, err := parseStringFloat(row[4])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse close price: %w", err)
	}

	baseVolume, err := parseStringFloat(row[5])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse base volume: %w", err)
	}

	quoteVolume, err := parseStringFloat(row[7])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse quote volume: %w", err)
	}

	trades, err := parseInt64(row[8])
	if err != nil {
		return domain.Candle{}, fmt.Errorf("parse trades: %w", err)
	}

	return domain.Candle{
		OpenTime:    openTime,
		CloseTime:   closeTime,
		OpenPrice:   openPrice,
		HighPrice:   highPrice,
		LowPrice:    lowPrice,
		ClosePrice:  closePrice,
		BaseVolume:  baseVolume,
		QuoteVolume: quoteVolume,
		Trades:      trades,
	}, nil
}

func parseUnixMilliseconds(value any) (time.Time, error) {
	switch typed := value.(type) {
	case float64:
		return time.UnixMilli(int64(typed)).UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time type %T", value)
	}
}

func parseStringFloat(value any) (float64, error) {
	text, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("unsupported float type %T", value)
	}

	return strconv.ParseFloat(text, 64)
}

func parseInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), nil
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}
