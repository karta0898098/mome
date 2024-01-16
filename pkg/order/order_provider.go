package order

import (
	"context"
)

// Provider define order service layer
type Provider interface {
	// Start process trade info
	Start(ctx context.Context)
	// SubmitOrder Submit order to order matching engine
	// Trade history will send by MQ when successful matching
	SubmitOrder(ctx context.Context, order Order) (err error)
	// ListAllAsks ist all asks orders. include Limit and Market orders
	ListAllAsks(ctx context.Context, symbol string) (orders []Order, err error)
	// ListAllBids List all bids orders include Limit and Market orders
	ListAllBids(ctx context.Context, symbol string) (orders []Order, err error)
}
