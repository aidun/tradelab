package domain

type Market struct {
	ID          string
	Symbol      string
	BaseAsset   string
	QuoteAsset  string
	MinNotional float64
	Exchange    string
}
