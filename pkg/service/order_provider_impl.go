package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/karta0898098/mome/pkg/order"
)

const (
	MinQty = 1
)

// check OrderProviderImpl is implement order.Provider
var _ order.Provider = &OrderProviderImpl{}

// OrderProviderImpl is implement for order.Provider
type OrderProviderImpl struct {
	OrderBooks map[string]*order.OrderBook
	OrderRepo  order.Repository
}

func NewOrderProviderImpl(orderBooks map[string]*order.OrderBook, orderRepo order.Repository) *OrderProviderImpl {
	return &OrderProviderImpl{
		OrderBooks: orderBooks,
		OrderRepo:  orderRepo,
	}
}

func (srv *OrderProviderImpl) Start(ctx context.Context) {
	go srv.processTradeEvents(ctx)
}

// SubmitOrder is implemented for order.Provider
func (srv *OrderProviderImpl) SubmitOrder(ctx context.Context, o order.Order) (err error) {
	orderBook, ok := srv.OrderBooks[o.TickerSymbol]
	if !ok {
		return fmt.Errorf("failed to submit order ticker symbol %s %w", o.TickerSymbol, order.ErrInvalidTickerSymbol)
	}

	// validate order book
	if o.Qty <= MinQty {
		return fmt.Errorf("failed to submit order qty %v %w", o.Qty, order.ErrInvalidQty)
	}

	if o.Kind == order.KindMarket && !o.Price.IsZero() {
		return fmt.Errorf("failed to submit order %w", order.ErrInvalidMarketPrice)
	}

	if o.Kind == order.KindLimit && o.Price.IsZero() {
		return fmt.Errorf("failed to submit order %w", order.ErrInvalidLimitPrice)
	}

	if o.Params.Is(order.ConditionStop) && o.StopPrice.IsZero() {
		return order.ErrInvalidStopPrice
	}

	// add order to match algorithm async
	logger := log.Ctx(ctx)
	logger.
		Info().
		Interface("order", o).
		Msg("add order to order books")
	_, err = orderBook.Add(ctx, o)
	if err != nil {
		logger.
			Error().
			Err(err).
			Msg("failed to add order to order book")
	}

	return nil
}

func (srv *OrderProviderImpl) processTradeEvents(ctx context.Context) {
LOOP:
	for {
		for _, book := range srv.OrderBooks {
			book := book
			go func(orderBook *order.OrderBook) {
				event := <-orderBook.TradeEvents
				fmt.Printf("save trade evetns %#v\n \n", event)
			}(book)
		}

		select {
		case <-ctx.Done():
			break LOOP
		}
	}
}
