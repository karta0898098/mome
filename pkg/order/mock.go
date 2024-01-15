package order

import "context"

type NopRepository struct {
}

func (n NopRepository) FindOrder(ctx context.Context, id string) (order *Order, err error) {
	return nil, err
}

func (n NopRepository) SaveOrder(ctx context.Context, order *Order) (err error) {
	return err
}
