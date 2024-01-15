package order

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/stretchr/testify/suite"
)

const instrument = "TEST"

type orderBookTestSuite struct {
	suite.Suite

	ob *OrderBook
}

func TestRun(t *testing.T) {
	suite.Run(t, new(orderBookTestSuite))
}

func (suite *orderBookTestSuite) SetupTest() {
	suite.ob = NewOrderBook(instrument, *apd.New(2025, -2), &NopRepository{})
}

func (suite *orderBookTestSuite) createOrder(id string, oType Kind, params Condition, qty int64, price, stopPrice apd.Decimal, side Side) Order {
	return Order{
		ID:           id,
		TickerSymbol: instrument,
		CustomerID:   xid.New().String(),
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

func (suite *orderBookTestSuite) TestOrderBook_MarketReject() {
	ob := suite.ob
	ctx := context.Background()

	matched, err := ob.Add(ctx, createOrder(
		"1",
		KindMarket,
		0,
		5,
		apd.Decimal{},
		apd.Decimal{},
		SideBuy,
	))
	suite.NoError(err)
	suite.False(matched, "expected no match for market order, got a match")
}

func (suite *orderBookTestSuite) TestOrderBook_MarketToLimit() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
		{
			matched: true,
			order:   createOrder("2", KindMarket, 0, 2, apd.Decimal{}, apd.Decimal{}, SideSell),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}
}

func (suite *orderBookTestSuite) TestOrderBook_LimitToMarket() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindMarket, 0, 2, apd.Decimal{}, apd.Decimal{}, SideSell),
		},
		{
			matched: true,
			order:   createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	trade := <-ob.TradeEvents

	suite.NotEmpty(trade)
	suite.Equal(0, ob.orders.Asks.Len())
	suite.Equal(1, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_No_Match() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 2, *apd.New(2025, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: false,
			order:   createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(1, ob.orders.Asks.Len())
	suite.Equal(1, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_Match() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: true,
			order:   createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(0, ob.orders.Asks.Len())
	suite.Equal(1, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_Match_FullQty() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: true,
			order:   createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(0, ob.orders.Asks.Len())
	suite.Equal(0, ob.orders.Bids.Len())
	suite.NotEqual(0, ob.activeOrders)
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_First_AON_Reject() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, ConditionAON, 5, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: false,
			order:   createOrder("2", KindLimit, 0, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_Second_AON_Reject() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: false,
			order:   createOrder("2", KindLimit, ConditionAON, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(1, ob.orders.Asks.Len())
	suite.Equal(1, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_First_IOC_Reject() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, ConditionIOC, 3, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: false,
			order:   createOrder("2", KindLimit, 0, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(0, ob.orders.Asks.Len())
	suite.Equal(1, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Limit_To_Limit_Second_IOC() {
	ob := suite.ob
	ctx := context.Background()

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 3, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: true,
			order:   createOrder("2", KindLimit, ConditionIOC, 2, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	suite.Equal(1, ob.orders.Asks.Len())
	suite.Equal(0, ob.orders.Bids.Len())
}

func (suite *orderBookTestSuite) TestOrderBook_Add_Bids() {
	// test order sorting
	ob := suite.ob
	t := suite.T()

	BaseContext := apd.Context{
		Precision:   0,               // no rounding
		MaxExponent: apd.MaxExponent, // up to 10^5 exponent
		MinExponent: apd.MinExponent, // support only 4 decimal places
		Traps:       apd.DefaultTraps,
	}
	ctx := context.Background()

	type orderData struct {
		ID        string
		Kind      Kind
		Params    Condition
		Qty       int64
		Price     apd.Decimal
		StopPrice apd.Decimal
		Side      Side
	}

	data := []orderData{
		{"0", KindLimit, 0, 5, *apd.New(2010, -2), apd.Decimal{}, SideBuy},
		{"1", KindMarket, ConditionAON, 11, apd.Decimal{}, apd.Decimal{}, SideBuy},
		{"2", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideBuy},
		{"3", KindLimit, 0, 2, *apd.New(2065, -2), apd.Decimal{}, SideBuy},
		{"4", KindMarket, 0, 4, apd.Decimal{}, apd.Decimal{}, SideBuy},
	}
	sorted := []int{1, 4, 3, 0, 2}

	for _, d := range data {
		_, _ = ob.Add(ctx, createOrder(d.ID, d.Kind, d.Params, d.Qty, d.Price, d.StopPrice, d.Side))
	}

	i := 0
	for iter := ob.orders.Bids.Iterator(); iter.Valid(); iter.Next() {
		order := ob.activeOrders[iter.Key().ID]
		expectedData := data[sorted[i]]

		var priceEq, stopPriceEq apd.Decimal
		_, err := BaseContext.Cmp(&priceEq, &expectedData.Price, &order.Price)
		suite.NoError(err)

		_, err = BaseContext.Cmp(&stopPriceEq, &expectedData.StopPrice, &order.StopPrice)
		suite.NoError(err)

		suite.Equal(expectedData.ID, order.ID)
		suite.Equal(expectedData.Kind, order.Kind)
		suite.Equal(expectedData.Params, order.Params)
		suite.Equal(expectedData.Qty, order.Qty)
		suite.Equal(expectedData.Side, order.Side)
		i += 1
		t.Logf("%+v", order)
	}
}

func (suite *orderBookTestSuite) TestOrderBook_Add_Asks() {
	// test order sorting
	ob := suite.ob
	t := suite.T()

	BaseContext := apd.Context{
		Precision:   0,               // no rounding
		MaxExponent: apd.MaxExponent, // up to 10^5 exponent
		MinExponent: apd.MinExponent, // support only 4 decimal places
		Traps:       apd.DefaultTraps,
	}
	ctx := context.Background()

	type orderData struct {
		ID        string
		Kind      Kind
		Params    Condition
		Qty       int64
		Price     apd.Decimal
		StopPrice apd.Decimal
		Side      Side
	}

	data := [...]orderData{
		{"0", KindLimit, 0, 7, *apd.New(2000, -2), apd.Decimal{}, SideSell},
		{"1", KindLimit, 0, 2, *apd.New(2013, -2), apd.Decimal{}, SideSell},
		{"2", KindLimit, 0, 8, *apd.New(2000, -2), apd.Decimal{}, SideSell},
		{"3", KindMarket, 0, 9, apd.Decimal{}, apd.Decimal{}, SideSell},
		{"4", KindLimit, 0, 3, *apd.New(2055, -2), apd.Decimal{}, SideSell},
	}

	sorted := []int{3, 0, 2, 1, 4}

	for _, d := range data {
		_, _ = ob.Add(ctx, createOrder(d.ID, d.Kind, d.Params, d.Qty, d.Price, d.StopPrice, d.Side))
	}

	i := 0
	for iter := ob.orders.Asks.Iterator(); iter.Valid(); iter.Next() {
		order := ob.activeOrders[iter.Key().ID]
		expectedData := data[sorted[i]]

		var priceEq, stopPriceEq apd.Decimal
		_, err := BaseContext.Cmp(&priceEq, &expectedData.Price, &order.Price)
		suite.NoError(err)

		_, err = BaseContext.Cmp(&stopPriceEq, &expectedData.StopPrice, &order.StopPrice)
		suite.NoError(err)

		suite.Equal(expectedData.ID, order.ID)
		suite.Equal(expectedData.Kind, order.Kind)
		suite.Equal(expectedData.Params, order.Params)
		suite.Equal(expectedData.Qty, order.Qty)
		suite.Equal(expectedData.Side, order.Side)
		i += 1
		t.Logf("%+v", order)
	}
}

func (suite *orderBookTestSuite) TestOrderBook_Add_MarketPrice_Change() {

	ob := suite.ob
	ctx := context.Background()

	BaseContext := apd.Context{
		Precision:   0,               // no rounding
		MaxExponent: apd.MaxExponent, // up to 10^5 exponent
		MinExponent: apd.MinExponent, // support only 4 decimal places
		Traps:       apd.DefaultTraps,
	}

	tests := []struct {
		matched bool
		order   Order
	}{
		{
			matched: false,
			order:   createOrder("1", KindLimit, 0, 2, *apd.New(2010, -2), apd.Decimal{}, SideSell),
		},
		{
			matched: true,
			order:   createOrder("2", KindLimit, 0, 5, *apd.New(2012, -2), apd.Decimal{}, SideBuy),
		},
	}
	for _, tt := range tests {
		matched, err := ob.Add(ctx, tt.order)
		suite.NoError(err)
		suite.Equal(tt.matched, matched)
	}

	var eq apd.Decimal
	_, err := BaseContext.Cmp(&eq, &ob.marketPrice, apd.New(2012, -2))
	suite.NoError(err)
}
