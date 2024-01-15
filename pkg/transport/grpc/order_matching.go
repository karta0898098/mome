package grpc

import (
	"context"

	"github.com/cockroachdb/apd"

	pb "github.com/karta0898098/mome/pb/order"
	"github.com/karta0898098/mome/pkg/order"
)

var _ pb.OrderMatchingServiceServer = &OrderMatchingHandler{}

type OrderMatchingHandler struct {
	provider order.Provider
}

func (h *OrderMatchingHandler) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderReply, error) {
	o, err := order.NewOrder(
		req.Symbol,
		req.CustomerID,
		order.Kind(req.Kind),
		order.Condition(req.Params),
		req.Quantity,
		*apd.New(req.Price.Coefficient, req.Price.Exponent),
		order.Side(req.Side),
	)
	if err != nil {
		return nil, err
	}

	err = h.provider.SubmitOrder(ctx, o)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitOrderReply{
		OrderID:        o.ID,
		CreatedAtMilli: o.CreatedAt.UnixMilli(),
	}, nil
}
