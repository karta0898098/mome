package order

import "context"

// Repository define order repository
type Repository interface {
	// FindOrder find order by id
	FindOrder(ctx context.Context, id string) (order *Order, err error)

	// SaveOrder save order
	SaveOrder(ctx context.Context, order *Order) (err error)
}
