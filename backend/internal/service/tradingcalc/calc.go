package tradingcalc

import (
	"math"
	"sort"

	"github.com/aidun/tradelab/backend/internal/domain"
)

type lot struct {
	quantity float64
	price    float64
}

type marketState struct {
	marketID     string
	marketSymbol string
	baseAsset    string
	quoteAsset   string
	openedAt     string

	averageQty      float64
	averagePrice    float64
	averageRealized float64

	fifoLots       []lot
	fifoRealized   float64
	lastOrderPrice float64
}

type AnalysisResult struct {
	Orders     []domain.Order
	Positions  []domain.Position
	Realized   float64
	Unrealized float64
}

func AnalyzeOrders(orders []domain.Order, mode domain.AccountingMode, currentPrices map[string]float64) AnalysisResult {
	if len(orders) == 0 {
		return AnalysisResult{Orders: []domain.Order{}, Positions: []domain.Position{}}
	}

	sortedOrders := append([]domain.Order(nil), orders...)
	sort.SliceStable(sortedOrders, func(i, j int) bool {
		return sortedOrders[i].CreatedAt.Before(sortedOrders[j].CreatedAt)
	})

	states := map[string]*marketState{}
	annotated := make([]domain.Order, 0, len(sortedOrders))

	for _, order := range sortedOrders {
		state := states[order.MarketSymbol]
		if state == nil {
			state = &marketState{
				marketID:       order.MarketID,
				marketSymbol:   order.MarketSymbol,
				baseAsset:      order.BaseAsset,
				quoteAsset:     order.QuoteAsset,
				openedAt:       order.CreatedAt.Format(timeLayout),
				lastOrderPrice: order.ExpectedPrice,
			}
			states[order.MarketSymbol] = state
		}
		state.lastOrderPrice = order.ExpectedPrice

		annotatedOrder := order
		switch order.Side {
		case domain.OrderSideBuy:
			applyAverageBuy(state, order.BaseQuantity, order.ExpectedPrice)
			applyFIFOBuy(state, order.BaseQuantity, order.ExpectedPrice)
		case domain.OrderSideSell:
			annotatedOrder.RealizedPnL = applySell(state, order.BaseQuantity, order.ExpectedPrice, mode)
		}

		annotatedOrder.PositionAfter = state.averageQty
		annotated = append(annotated, annotatedOrder)
	}

	positions := make([]domain.Position, 0, len(states))
	var realizedTotal float64
	var unrealizedTotal float64

	for _, state := range states {
		if state.averageQty <= epsilon {
			realizedTotal += realizedForMode(state, mode)
			continue
		}

		currentPrice := currentPrices[state.marketSymbol]
		if currentPrice <= 0 {
			currentPrice = state.lastOrderPrice
		}

		entryPriceAvg, costBasisValue := costBasisForMode(state, mode)
		positionValue := state.averageQty * currentPrice
		unrealized := positionValue - costBasisValue
		realized := realizedForMode(state, mode)

		positions = append(positions, domain.Position{
			ID:             state.marketID,
			MarketID:       state.marketID,
			MarketSymbol:   state.marketSymbol,
			BaseAsset:      state.baseAsset,
			QuoteAsset:     state.quoteAsset,
			Status:         "open",
			OpenQuantity:   state.averageQty,
			EntryQuantity:  state.averageQty,
			EntryPriceAvg:  entryPriceAvg,
			CurrentPrice:   currentPrice,
			CostBasisValue: costBasisValue,
			PositionValue:  positionValue,
			RealizedPnL:    realized,
			UnrealizedPnL:  unrealized,
		})

		realizedTotal += realized
		unrealizedTotal += unrealized
	}

	sort.SliceStable(positions, func(i, j int) bool {
		return positions[i].MarketSymbol < positions[j].MarketSymbol
	})

	sort.SliceStable(annotated, func(i, j int) bool {
		return annotated[i].CreatedAt.After(annotated[j].CreatedAt)
	})

	return AnalysisResult{
		Orders:     annotated,
		Positions:  positions,
		Realized:   realizedTotal,
		Unrealized: unrealizedTotal,
	}
}

const (
	epsilon    = 0.00000001
	timeLayout = "2006-01-02T15:04:05Z07:00"
)

func applyAverageBuy(state *marketState, quantity float64, price float64) {
	totalCost := (state.averageQty * state.averagePrice) + (quantity * price)
	state.averageQty += quantity
	if state.averageQty <= epsilon {
		state.averageQty = 0
		state.averagePrice = 0
		return
	}
	state.averagePrice = totalCost / state.averageQty
}

func applyFIFOBuy(state *marketState, quantity float64, price float64) {
	state.fifoLots = append(state.fifoLots, lot{quantity: quantity, price: price})
}

func applySell(state *marketState, quantity float64, price float64, mode domain.AccountingMode) float64 {
	avgRealized := 0.0
	if state.averageQty > epsilon {
		avgRealized = (price - state.averagePrice) * quantity
		state.averageQty = math.Max(0, state.averageQty-quantity)
		if state.averageQty <= epsilon {
			state.averageQty = 0
			state.averagePrice = 0
		}
		state.averageRealized += avgRealized
	}

	fifoRealized := 0.0
	remaining := quantity
	updatedLots := make([]lot, 0, len(state.fifoLots))
	for _, currentLot := range state.fifoLots {
		if remaining <= epsilon {
			updatedLots = append(updatedLots, currentLot)
			continue
		}

		consumed := math.Min(currentLot.quantity, remaining)
		fifoRealized += (price - currentLot.price) * consumed
		remaining -= consumed
		left := currentLot.quantity - consumed
		if left > epsilon {
			updatedLots = append(updatedLots, lot{quantity: left, price: currentLot.price})
		}
	}
	state.fifoLots = updatedLots
	state.fifoRealized += fifoRealized

	switch mode {
	case domain.AccountingModeFIFO, domain.AccountingModeHybrid:
		return fifoRealized
	default:
		return avgRealized
	}
}

func costBasisForMode(state *marketState, mode domain.AccountingMode) (float64, float64) {
	switch mode {
	case domain.AccountingModeFIFO:
		var quantity float64
		var cost float64
		for _, currentLot := range state.fifoLots {
			quantity += currentLot.quantity
			cost += currentLot.quantity * currentLot.price
		}
		if quantity <= epsilon {
			return 0, 0
		}
		return cost / quantity, cost
	default:
		return state.averagePrice, state.averageQty * state.averagePrice
	}
}

func realizedForMode(state *marketState, mode domain.AccountingMode) float64 {
	switch mode {
	case domain.AccountingModeFIFO, domain.AccountingModeHybrid:
		return state.fifoRealized
	default:
		return state.averageRealized
	}
}
