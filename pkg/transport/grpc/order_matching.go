package grpc

import (
	"context"

	"github.com/cockroachdb/apd"
	"github.com/rs/zerolog/log"

	pb "github.com/karta0898098/mome/pb/order"
	"github.com/karta0898098/mome/pkg/order"
)

var _ pb.OrderMatchingServiceServer = &OrderMatchingHandler{}

type OrderMatchingHandler struct {
	provider order.Provider
}

func NewOrderMatchingHandler(provider order.Provider) *OrderMatchingHandler {
	return &OrderMatchingHandler{provider: provider}
}

func (h *OrderMatchingHandler) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderReply, error) {
	logger := log.Ctx(ctx)
	logger.Debug().Interface("req", req).Msg("debug...")

	price := apd.New(0, 0)
	if req.Price != nil {
		price = apd.New(req.Price.Coefficient, req.Price.Exponent)
	}

	stopPrice := apd.New(0, 0)
	if req.StopPrice != nil {
		stopPrice = apd.New(req.StopPrice.Coefficient, req.StopPrice.Exponent)
	}

	var params order.Condition
	switch req.Params {
	case pb.OrderParams_ORDER_PARAMS_STOP:
		params |= order.ConditionStop
	case pb.OrderParams_ORDER_PARAMS_AON:
		params |= order.ConditionAON
	case pb.OrderParams_ORDER_PARAMS_IOC:
		params |= order.ConditionIOC
	case pb.OrderParams_ORDER_PARAMS_FOK:
		params |= order.ConditionFOK
	case pb.OrderParams_ORDER_PARAMS_GTC:
		params |= order.ConditionGTC
	case pb.OrderParams_ORDER_PARAMS_GFD:
		params |= order.ConditionGFD
	case pb.OrderParams_ORDER_PARAMS_GTD:
		params |= order.ConditionGTD
	}

	o, err := order.NewOrder(
		req.Symbol,
		req.CustomerID,
		order.Kind(req.Kind),
		params,
		req.Quantity,
		price,
		stopPrice,
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

func (h *OrderMatchingHandler) ListAllAsks(ctx context.Context, req *pb.ListAllAsksRequest) (*pb.ListAllAskReply, error) {
	orders, err := h.provider.ListAllAsks(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	reply := &pb.ListAllAskReply{
		Orders: nil,
	}

	for _, o := range orders {
		o := o
		params := pb.OrderParams_value["ORDER_PARAMS_"+o.Params.String()]
		reply.Orders = append(reply.Orders, &pb.Order{
			ID:   o.ID,
			Kind: pb.OrderKind(o.Kind),
			Price: &pb.Price{
				Coefficient: o.Price.Coeff.Int64(),
				Exponent:    o.Price.Exponent,
			},
			StopPrice: &pb.Price{
				Coefficient: o.StopPrice.Coeff.Int64(),
				Exponent:    o.StopPrice.Exponent,
			},
			CreatedAtMilli: o.CreatedAt.UnixMilli(),
			Quantity:       o.Qty,
			FilledQuantity: o.FilledQty,
			Params:         pb.OrderParams(params),
		})
	}

	return reply, nil
}

func (h *OrderMatchingHandler) ListAllBids(ctx context.Context, req *pb.ListAllBidsRequest) (*pb.ListAllBidsReply, error) {
	orders, err := h.provider.ListAllBids(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	reply := &pb.ListAllBidsReply{
		Orders: nil,
	}

	for _, o := range orders {
		o := o
		params := pb.OrderParams_value["ORDER_PARAMS_"+o.Params.String()]
		reply.Orders = append(reply.Orders, &pb.Order{
			ID:   o.ID,
			Kind: pb.OrderKind(o.Kind),
			Price: &pb.Price{
				Coefficient: o.Price.Coeff.Int64(),
				Exponent:    o.Price.Exponent,
			},
			StopPrice: &pb.Price{
				Coefficient: o.StopPrice.Coeff.Int64(),
				Exponent:    o.StopPrice.Exponent,
			},
			CreatedAtMilli: o.CreatedAt.UnixMilli(),
			Quantity:       o.Qty,
			FilledQuantity: o.FilledQty,
			Params:         pb.OrderParams(params),
		})
	}

	return reply, nil
}
