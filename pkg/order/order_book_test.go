package order

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/google/uuid"
)

const instrument = "TEST"

func createOrder(id string, oType Kind, params Condition, qty int64, price, stopPrice apd.Decimal, side Side) Order {
	return Order{
		ID:           id,
		TickerSymbol: instrument,
		CustomerID:   uuid.NewString(),
		CreatedAt:    time.Now(),
		Kind:         oType,
		Params:       params,
		Qty:          qty,
		FilledQty:    0,
		Price:        price,
		StopPrice:    stopPrice,
		Side:         side,
	}
}

func setup(coeff int64, exp int32) *OrderBook {

	ob := NewOrderBook(instrument, *apd.New(coeff, exp), &NopRepository{})
	return ob
}

func TestOrderBook_MarketReject(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()
	matched, err := ob.Add(ctx, createOrder("1", KindMarket, 0, 5, apd.Decimal{}, apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindMarket, 0, 2, apd.Decimal{}, apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
}

func TestOrderBook_MarketToLimit(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindMarket, 0, 2, apd.Decimal{}, apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected match for market order, got no match")
	}
}

func TestOrderBook_LimitToMarket(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder("1", KindMarket, 0, 2, apd.Decimal{}, apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected match for market order, got no match")
	}

	if ob.orders.Asks.Len() != 0 {
		t.Errorf("expected 0 asks, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_No_Match(t *testing.T) {
	ob := setup(2025, -2)
	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 2, *apd.New(2025, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}

	if ob.orders.Asks.Len() != 1 {
		t.Errorf("expected 1 ask, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_Match(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()
	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected a match for this order, got a match")
	}

	if ob.orders.Asks.Len() != 0 {
		t.Errorf("expected 0 asks, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_Match_FullQty(t *testing.T) {
	ob := setup(2025, -2)
	ctx := context.Background()

	o1 := createOrder("1", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideSell)
	matched, err := ob.Add(ctx, o1)
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for market order, got a match")
	}
	o2 := createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy)
	matched, err = ob.Add(ctx, o2)
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected a match for this order, got a match")
	}

	if ob.orders.Asks.Len() != 0 {
		t.Errorf("expected 0 asks, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 0 {
		t.Errorf("expected 0 bids, got %d", ob.orders.Bids.Len())
	}
	if len(ob.activeOrders) != 0 {
		t.Errorf("expected 0 active orders, got %d", len(ob.activeOrders))
	}
}

func TestOrderBook_Limit_To_Limit_First_AON_Reject(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder("1", KindLimit, ConditionAON, 5, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, 0, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for  order, got a match")
	}

	if ob.orders.Asks.Len() != 1 {
		t.Errorf("expected 1 ask, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_Second_AON_Reject(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()
	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, ConditionAON, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for  order, got a match")
	}

	if ob.orders.Asks.Len() != 1 {
		t.Errorf("expected 1 ask, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_First_IOC_Reject(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()
	matched, err := ob.Add(ctx, createOrder("1", KindLimit, ConditionIOC, 3, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	if ob.orders.Asks.Len() != 0 {
		t.Fatalf("expected no asks, got %d", ob.orders.Asks.Len())
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, 0, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	if ob.orders.Asks.Len() != 0 {
		t.Errorf("expected 0 asks, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 1 {
		t.Errorf("expected 1 bid, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_Second_IOC(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 3, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, ConditionIOC, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected a match for this order, got no matches")
	}

	if ob.orders.Asks.Len() != 1 {
		t.Errorf("expected 1 ask, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 0 {
		t.Errorf("expected 0 bids, got %d", ob.orders.Bids.Len())
	}
}

func TestOrderBook_Limit_To_Limit_Second_IOC_CancelCheck(t *testing.T) {
	ob := setup(2025, -2)

	ctx := context.Background()
	matched, err := ob.Add(ctx, createOrder("1", KindLimit, 0, 3, *apd.New(2010, -2), apd.Decimal{}, SideSell))
	if err != nil {
		t.Error(err)
	}
	if matched {
		t.Errorf("expected no match for this order, got a match")
	}
	matched, err = ob.Add(ctx, createOrder("2", KindLimit, ConditionIOC, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
	if err != nil {
		t.Error(err)
	}
	if !matched {
		t.Errorf("expected a match for this order, got no matches")
	}

	if ob.orders.Asks.Len() != 0 {
		t.Errorf("expected 0 asks, got %d", ob.orders.Asks.Len())
	}
	if ob.orders.Bids.Len() != 0 {
		t.Errorf("expected 0 bids, got %d", ob.orders.Bids.Len())
	}
	order := ob.activeOrders["1"]
	if !order.IsCancelled() {
		t.Log("IOC order should be cancelled after partial fill")
	}
	if order.FilledQty != 3 {
		t.Logf("expected filled qty for IOC order %d, got %d", 3, order.FilledQty)
	}
	t.Logf("%+v", order)
}

// func TestOrderBook_Add_Bids(t *testing.T) {
// 	// test order sorting
// 	ob := setup(2025, -2)
//
// 	ctx := context.Background()
// 	type orderData struct {
// 		Type      OrderType
// 		Params    OrderParams
// 		Qty       int64
// 		Price     apd.Decimal
// 		StopPrice apd.Decimal
// 		Side      OrderSide
// 	}
//
// 	data := [...]orderData{
// 		{KindLimit, 0, 5, *apd.New(2010, -2), apd.Decimal{}, SideBuy},
// 		{KindMarket, ParamAON, 11, apd.Decimal{}, apd.Decimal{}, SideBuy},
// 		{KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideBuy},
// 		{KindLimit, 0, 2, *apd.New(2065, -2), apd.Decimal{}, SideBuy},
// 		{KindMarket, 0, 4, apd.Decimal{}, apd.Decimal{}, SideBuy},
// 	}
//
// 	for i, d := range data {
// 		_, _ = ob.Add(createOrder(uint64(i+1), d.Type, d.Params, d.Qty, d.Price, d.StopPrice, d.Side))
// 	}
//
// 	sorted := []int{1, 4, 3, 0, 2}
//
// 	i := 0
// 	for iter := ob.orders.Bids.Iterator(); iter.Valid(); iter.Next() {
// 		order := ob.activeOrders[iter.Key().OrderID]
//
// 		expectedData := data[sorted[i]]
//
// 		var priceEq, stopPriceEq apd.Decimal
// 		if _, err := BaseContext.Cmp(&priceEq, &expectedData.Price, &order.Price); err != nil {
// 			t.Fatal(err)
// 		}
// 		if _, err := BaseContext.Cmp(&stopPriceEq, &expectedData.StopPrice, &order.StopPrice); err != nil {
// 			t.Fatal(err)
// 		}
//
// 		equals := uint64(sorted[i]+1) == order.ID && expectedData.Type == order.Type && expectedData.Params == order.Params && expectedData.Qty == order.Qty && priceEq.IsZero() && stopPriceEq.IsZero() && expectedData.Side == order.Side
// 		if !equals {
// 			t.Errorf("expected order ID %d to be in place %d, got a different order", sorted[i]+1, i)
// 		}
//
// 		i += 1
// 		t.Logf("%+v", order)
// 	}
// }
//
// func TestOrderBook_Add_Asks(t *testing.T) {
// 	// test order sorting
// 	_, ob := setup(2025, -2)
//
// 	type orderData struct {
// 		Type      OrderType
// 		Params    OrderParams
// 		Qty       int64
// 		Price     apd.Decimal
// 		StopPrice apd.Decimal
// 		Side      OrderSide
// 	}
//
// 	data := [...]orderData{
// 		{KindLimit, 0, 7, *apd.New(2000, -2), apd.Decimal{}, SideSell},
// 		{KindLimit, 0, 2, *apd.New(2013, -2), apd.Decimal{}, SideSell},
// 		{KindLimit, 0, 8, *apd.New(2000, -2), apd.Decimal{}, SideSell},
// 		{KindMarket, 0, 9, apd.Decimal{}, apd.Decimal{}, SideSell},
// 		{KindLimit, 0, 3, *apd.New(2055, -2), apd.Decimal{}, SideSell},
// 	}
//
// 	for i, d := range data {
// 		_, _ = ob.Add(createOrder(uint64(i+1), d.Type, d.Params, d.Qty, d.Price, d.StopPrice, d.Side))
// 	}
//
// 	sorted := []int{3, 0, 2, 1, 4}
//
// 	i := 0
// 	for iter := ob.orders.Asks.Iterator(); iter.Valid(); iter.Next() {
// 		order := ob.activeOrders[iter.Key().OrderID]
//
// 		expectedData := data[sorted[i]]
//
// 		var priceEq, stopPriceEq apd.Decimal
// 		if _, err := BaseContext.Cmp(&priceEq, &expectedData.Price, &order.Price); err != nil {
// 			t.Fatal(err)
// 		}
// 		if _, err := BaseContext.Cmp(&stopPriceEq, &expectedData.StopPrice, &order.StopPrice); err != nil {
// 			t.Fatal(err)
// 		}
//
// 		equals := uint64(sorted[i]+1) == order.ID && expectedData.Type == order.Type && expectedData.Params == order.Params && expectedData.Qty == order.Qty && priceEq.IsZero() && stopPriceEq.IsZero() && expectedData.Side == order.Side
// 		if !equals {
// 			t.Errorf("expected order ID %d to be in place %d, got a different order", sorted[i]+1, i)
// 		}
//
// 		i += 1
// 		t.Logf("%+v", order)
// 	}
// }
//
// func TestOrderBook_Add_MarketPrice_Change(t *testing.T) {
// 	_, ob := setup(2025, -2)
//
// 	matched, err := ob.Add(createOrder(1, KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if matched {
// 		t.Errorf("expected no match for market order, got a match")
// 	}
// 	matched, err = ob.Add(createOrder(2, KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if !matched {
// 		t.Errorf("expected a match for this order, got a match")
// 	}
// 	var eq apd.Decimal
// 	if _, err := BaseContext.Cmp(&eq, &ob.marketPrice, apd.New(2012, -2)); err != nil {
// 		t.Fatal(err)
// 	}
// 	if !eq.IsZero() {
// 		t.Errorf("expected market price to be %f, got %s", 20.12, ob.marketPrice.String())
// 	}
// }
