package domain

import "time"

// Market describes a tradable symbol and its execution constraints.
type Market struct {
	ID          string  `json:"id"`
	Symbol      string  `json:"symbol"`
	BaseAsset   string  `json:"baseAsset"`
	QuoteAsset  string  `json:"quoteAsset"`
	MinNotional float64 `json:"minNotional"`
	Exchange    string  `json:"exchange"`
}

type Candle struct {
	OpenTime    time.Time `json:"openTime"`
	CloseTime   time.Time `json:"closeTime"`
	OpenPrice   float64   `json:"openPrice"`
	HighPrice   float64   `json:"highPrice"`
	LowPrice    float64   `json:"lowPrice"`
	ClosePrice  float64   `json:"closePrice"`
	BaseVolume  float64   `json:"baseVolume"`
	QuoteVolume float64   `json:"quoteVolume"`
	Trades      int64     `json:"trades"`
}

type MarketDataMeta struct {
	Source      string    `json:"source"`
	GeneratedAt time.Time `json:"generated_at"`
}

// CandleFeed bundles chart candles with metadata about feed freshness.
type CandleFeed struct {
	Candles []Candle
	Meta    MarketDataMeta
}
