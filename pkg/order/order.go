package order

import (
	"strings"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/rs/xid"
)

type Kind int8

const (
	KindMarket Kind = iota + 1
	KindLimit
)

func (o Kind) String() string {
	switch o {
	case KindMarket:
		return "Market"
	case KindLimit:
		return "Limit"
	default:
		return "invalid"
	}
}

type Condition int8

const (
	ConditionStop Condition = 0x1                         // stop order (has to have stop price set)
	ConditionAON  Condition = 0x2                         // all-or-nothing - complete fill or cancel
	ConditionIOC  Condition = 0x4                         // immediate-or-cancel - immediately fill what you can, cancel the rest
	ConditionFOK  Condition = ConditionIOC | ConditionAON // immediately try to fill the whole order
	ConditionGTC  Condition = 0x10                        // good-till-cancelled -  keep order active until manually cancelled
	ConditionGFD  Condition = 0x20                        // good-for-day keep order active until the end of the trading day
	ConditionGTD  Condition = 0x40                        // good-till-date - keep order active until the provided date (including the date)
)

func (o Condition) appendStr(hasPrefix bool, sb *strings.Builder, param Condition, value string) bool {
	if o.Is(param) {
		if hasPrefix {
			sb.WriteRune(' ')
		}
		sb.WriteString(value)
		return true
	}
	return hasPrefix
}

func (o Condition) String() string {
	var sb strings.Builder
	added := false
	added = o.appendStr(added, &sb, ConditionStop, "STOP")
	added = o.appendStr(added, &sb, ConditionFOK, "FOK")
	if !o.Is(ConditionFOK) {
		added = o.appendStr(added, &sb, ConditionAON, "AON")
		added = o.appendStr(added, &sb, ConditionIOC, "IOC")
	}
	return sb.String()
}

// Is returns true if a parameter value matches the provided parameters (if param is a subset of o)
// e.g. ParamFOK.Is(ParamAON) is true, ParamFOK.Is(ParamStop) is false. ParamAON.Is(ParamAON) is true.
func (o Condition) Is(param Condition) bool {
	return o&param == param
}

type Side uint

const (
	SideBuy Side = iota + 1
	SideSell
)

type Order struct {
	ID string
	// this ticker symbol
	TickerSymbol string
	// CreatedAt means this order arrive time
	CreatedAt time.Time
	// the customer id
	CustomerID string

	Kind      Kind      // order kind - market or limit
	Params    Condition // order parameters which change the way an order is stored and matched
	Qty       int64
	FilledQty int64       // currently filled quantity
	Price     apd.Decimal // used in limit orders
	StopPrice apd.Decimal // used in stop orders
	Side      Side        // determines whether an order is a bid (buy) or an ask (sell)
	Cancelled bool        // determines if an order is cancelled. A partially filled order can be cancelled.
}

func NewOrder(
	tickerSymbol string,
	customerID string,
	kind Kind,
	params Condition,
	qty int64,
	price *apd.Decimal,
	stopPrice *apd.Decimal,
	side Side,
) (Order, error) {
	return Order{
		ID:           xid.New().String(),
		TickerSymbol: tickerSymbol,
		CreatedAt:    time.Now().UTC(),
		CustomerID:   customerID,
		Kind:         kind,
		Params:       params,
		Qty:          qty,
		FilledQty:    0,
		Price:        *price,
		StopPrice:    *stopPrice,
		Side:         side,
		Cancelled:    false,
	}, nil
}

func (o *Order) IsCancelled() bool {
	return o.Cancelled
}

func (o *Order) IsFilled() bool {
	return o.Qty-o.FilledQty == 0
}

func (o *Order) IsBid() bool {
	return o.Side == SideBuy
}

func (o *Order) IsAsk() bool {
	return o.Side == SideSell
}

func (o *Order) Cancel() {
	o.Cancelled = true
}

func (o *Order) UnfilledQty() int64 {
	return o.Qty - o.FilledQty
}

type OrderTracker struct {
	ID        string
	Kind      Kind
	Price     float64
	Side      Side
	Timestamp int64 // nanoseconds since Epoch
}
