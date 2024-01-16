package order

import "errors"

var (
	ErrInvalidQty          = errors.New("invalid quantity provided")
	ErrInvalidTickerSymbol = errors.New("invalid ticker symbol")
	ErrInvalidMarketPrice  = errors.New("price has to be zero for market orders")
	ErrInvalidLimitPrice   = errors.New("price has to be set for limit orders")
	ErrInvalidStopPrice    = errors.New("stop price has to be set for a stop order")
	ErrInternal            = errors.New("internal error")
)
