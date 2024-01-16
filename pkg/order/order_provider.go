package order

import (
	"context"
)

// Provider define order service layer
type Provider interface {
	Start(ctx context.Context)
	// SubmitOrder submit new order
	SubmitOrder(ctx context.Context, order Order) (err error)

	ListAllAsks(ctx context.Context, symbol string) (orders []Order, err error)

	ListAllBids(ctx context.Context, symbol string) (orders []Order, err error)
}
