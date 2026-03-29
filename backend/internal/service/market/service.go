package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aidun/tradelab/backend/internal/domain"
	"github.com/aidun/tradelab/backend/internal/store"
)

type Service struct {
	markets           store.MarketRepository
	marketDataBaseURL string
	client            *http.Client
}

func NewService(markets store.MarketRepository, marketDataBaseURL string) *Service {
	return &Service{
		markets:           markets,
		marketDataBaseURL: strings.TrimRight(marketDataBaseURL, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) List(ctx context.Context) ([]domain.Market, error) {
	return s.markets.List(ctx)
}

func (s *Service) ListCandles(ctx context.Context, marketSymbol string, interval string, limit int) ([]domain.Candle, error) {
	if interval == "" {
		interval = "1h"
	}

	if limit <= 0 || limit > 200 {
		limit = 48
	}

	market, err := s.markets.GetBySymbol(ctx, marketSymbol)
	if err != nil {
		return nil, fmt.Errorf("get market: %w", err)
	}

	query := url.Values{}
	query.Set("symbol", strings.ReplaceAll(market.Symbol, "/", ""))
	query.Set("interval", interval)
	query.Set("limit", strconv.Itoa(limit))

	endpoint := fmt.Sprintf("%s/api/v3/uiKlines?%s", s.marketDataBaseURL, query.Encode())
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create market data request: %w", err)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request market data: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market data request failed with status %d", response.StatusCode)
	}

	var payload [][]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode market data: %w", err)
	}

	candles := make([]domain.Candle, 0, len(payload))
	for _, row := range payload {
		candle, err := parseCandle(row)
		if err != nil {
			return nil, err
		}
		candles = append(candles, candle)
	}

	return candles, nil
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
