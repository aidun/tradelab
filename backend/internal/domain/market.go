package domain

import "time"

type Market struct {
	ID          string
	Symbol      string
	BaseAsset   string
	QuoteAsset  string
	MinNotional float64
	Exchange    string
}

type Candle struct {
	OpenTime    time.Time
	CloseTime   time.Time
	OpenPrice   float64
	HighPrice   float64
	LowPrice    float64
	ClosePrice  float64
	BaseVolume  float64
	QuoteVolume float64
	Trades      int64
}

type MarketDataMeta struct {
	Source      string    `json:"source"`
	GeneratedAt time.Time `json:"generated_at"`
}

type CandleFeed struct {
	Candles []Candle
	Meta    MarketDataMeta
}
