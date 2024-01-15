package order

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/google/uuid"
	"github.com/igrmk/treemap/v2"
)

const (
	MinQty = 1
)

type OrderBook struct {
	TickerSymbol string

	marketPrice      apd.Decimal
	marketPriceMutex sync.RWMutex

	orderRepo    Repository       // persistent order storage
	activeOrders map[string]Order // quick order retrieval by ID
	orderMutex   sync.RWMutex

	orders     *Set // contains all orders
	stopOrders *Set

	TradeEvents chan EventTradeSuccess
}

type EventTradeSuccess struct {
	ID           string
	Buyer        string
	Seller       string
	TickerSymbol string
	Qty          int64
	Price        apd.Decimal
	Total        apd.Decimal
	Timestamp    time.Time

	BidOrderID string
	AskOrderID string
}

func NewOrderBook(symbol string, marketPrice apd.Decimal, orderRepo Repository) *OrderBook {
	// bid need price high
	bid := newComparator(true)
	// asker need price lower
	ask := newComparator(false)

	stopBidLess := newStopComparator(false)
	stopAskLess := newStopComparator(true)

	return &OrderBook{
		TickerSymbol: symbol,
		marketPrice:  marketPrice,
		orderRepo:    orderRepo,
		activeOrders: make(map[string]Order),
		orders:       NewOrderSet(bid, ask),
		stopOrders:   NewOrderSet(stopBidLess, stopAskLess),
		TradeEvents:  make(chan EventTradeSuccess, 10000),
	}
}

// Get all bids ordered the same way they are matched.
func (o *OrderBook) GetBids() []Order {
	o.orderMutex.RLock()
	defer o.orderMutex.RUnlock()
	orders := make([]Order, 0, o.orders.Len(SideBuy))
	for iter := o.orders.Iterator(SideBuy); iter.Valid(); iter.Next() {
		orders = append(orders, o.activeOrders[iter.Key().ID])
	}

	return orders
}

// Get all asks ordered the same way they are matched.
func (o *OrderBook) GetAsks() []Order {
	o.orderMutex.RLock()
	defer o.orderMutex.RUnlock()
	orders := make([]Order, 0, o.orders.Len(SideSell))
	for iter := o.orders.Iterator(SideSell); iter.Valid(); iter.Next() {
		orders = append(orders, o.activeOrders[iter.Key().ID])
	}
	return orders
}

// Get all stop bids.
func (o *OrderBook) GetStopBids() []Order {
	o.orderMutex.RLock()
	defer o.orderMutex.RUnlock()
	orders := make([]Order, 0, o.stopOrders.Len(SideBuy))
	for iter := o.stopOrders.Iterator(SideBuy); iter.Valid(); iter.Next() {
		orders = append(orders, o.activeOrders[iter.Key().ID])
	}

	return orders
}

// Get all stop asks.
func (o *OrderBook) GetStopAsks() []Order {
	o.orderMutex.RLock()
	defer o.orderMutex.RUnlock()
	orders := make([]Order, 0, o.stopOrders.Len(SideSell))
	for iter := o.stopOrders.Iterator(SideSell); iter.Valid(); iter.Next() {
		orders = append(orders, o.activeOrders[iter.Key().ID])
	}
	return orders
}

// Get a market price.
func (o *OrderBook) MarketPrice() apd.Decimal {
	o.marketPriceMutex.RLock()
	defer o.marketPriceMutex.RUnlock()
	return o.marketPrice
}

// Set a market price.
func (o *OrderBook) SetMarketPrice(ctx context.Context, price apd.Decimal, fPrice float64) {
	o.marketPriceMutex.Lock()
	o.marketPrice = price
	o.marketPriceMutex.Unlock()

	bids := o.stopOrders.FindAllBidsBelow(fPrice)
	o.addOrders(ctx, bids)
	asks := o.stopOrders.FindAllAsksAbove(fPrice)
	o.addOrders(ctx, asks)
}

func (o *OrderBook) addOrders(ctx context.Context, trackers []OrderTracker) {
	for _, bid := range trackers {
		order, ok := o.findActiveOrder(bid.ID)
		o.stopOrders.Remove(bid.ID)
		if !ok {
			panic(fmt.Errorf("order with ID %s not found", bid.ID))
		}
		if _, err := o.submit(ctx, order, bid); err != nil {
			log.Println(err) // todo: better handling of these events
		}
	}
}

func (o *OrderBook) findActiveOrder(id string) (Order, bool) {
	o.orderMutex.RLock()
	defer o.orderMutex.RUnlock()
	order, ok := o.activeOrders[id]

	return order, ok
}

// Insert an order in activeOrders map.
func (o *OrderBook) setActiveOrder(order Order) error {
	o.orderMutex.Lock()
	defer o.orderMutex.Unlock()
	if _, ok := o.activeOrders[order.ID]; ok {
		return fmt.Errorf("order with ID %s already exists", order.ID)
	}
	o.activeOrders[order.ID] = order
	return nil
}

// Add an order to books - make it matchable against other orders.
func (o *OrderBook) addToBooks(tracker OrderTracker) {
	o.orderMutex.Lock()
	o.orders.Add(tracker) // enter pointer to the tree
	o.orderMutex.Unlock()
}

func (o *OrderBook) storeOrder(ctx context.Context, order Order) error {
	if err := o.setActiveOrder(order); err != nil {
		o.orders.Remove(order.ID)
		return err
	}
	return o.orderRepo.SaveOrder(ctx, &order)
}

// Update an active order.
func (o *OrderBook) updateActiveOrder(ctx context.Context, order Order) error {
	o.orderMutex.Lock()
	defer o.orderMutex.Unlock()
	if _, ok := o.activeOrders[order.ID]; !ok {
		return fmt.Errorf("order with ID %s hasn't yet been saved", order.ID)
	}
	o.activeOrders[order.ID] = order
	return o.orderRepo.SaveOrder(ctx, &order)
}

// Removes an order from books - removes it from possible matches.
func (o *OrderBook) removeFromBooks(ctx context.Context, orderID string) {
	order, ok := o.findActiveOrder(orderID)
	if !ok {
		return
	}
	if err := o.orderRepo.SaveOrder(ctx, &order); err != nil { // ensure we store the latest order data
		log.Printf("cannot save the order %+v to the repo - repository data might be inconsistent\n", order.ID)
	}

	o.orderMutex.Lock()
	o.orders.Remove(orderID)
	delete(o.activeOrders, orderID) // remove an active order
	o.orderMutex.Unlock()
}

// Cancel an order.
func (o *OrderBook) Cancel(ctx context.Context, id string) error {
	o.orderMutex.RLock()
	order, ok := o.activeOrders[id]
	o.orderMutex.RUnlock()

	if !ok {
		return nil
	}
	order.Cancel()
	return o.updateActiveOrder(ctx, order) // todo: remove from active orders
}

// Add a new order. Order can be matched immediately or later (or never), depending on order parameters and order type.
// Returns true if order was matched (partially or fully), false otherwise.
func (o *OrderBook) Add(ctx context.Context, order Order) (bool, error) {
	if order.Qty <= MinQty { // check the qty
		return false, ErrInvalidQty
	}
	if order.Kind == KindMarket && !order.Price.IsZero() {
		return false, ErrInvalidMarketPrice
	}
	if order.Kind == KindLimit && order.Price.IsZero() {
		return false, ErrInvalidLimitPrice
	}
	if order.Params.Is(ConditionStop) && order.StopPrice.IsZero() {
		return false, ErrInvalidStopPrice
	}

	orderPrice, err := order.Price.Float64()
	if err != nil {
		return false, err
	}

	tracker := OrderTracker{
		ID:        order.ID,
		Kind:      order.Kind,
		Price:     orderPrice,
		Side:      order.Side,
		Timestamp: order.CreatedAt.UnixNano(),
	}

	if order.Params.Is(ConditionStop) {
		marketPrice := o.MarketPrice()

		orderStopPrice, err := order.StopPrice.Float64()
		if err != nil {
			return false, err
		}

		tracker := OrderTracker{
			ID:        order.ID,
			Kind:      order.Kind,
			Price:     orderStopPrice,
			Side:      order.Side,
			Timestamp: order.CreatedAt.UnixNano(),
		}

		switch order.Side {
		case SideBuy:
			// if market price is lower than the bid stop price add as a stop order
			// otherwise process immediately
			if marketPrice.Cmp(&order.StopPrice) < 0 {
				o.stopOrders.Add(tracker)
				if err := o.storeOrder(ctx, order); err != nil {
					return false, err
				}
				return false, nil
			}
		case SideSell:
			// if market price is higher than the ask stop price add as a stop order
			// otherwise proces immediately
			if marketPrice.Cmp(&order.StopPrice) > 0 {
				o.stopOrders.Add(tracker)
				if err := o.storeOrder(ctx, order); err != nil {
					return false, err
				}
				return false, nil
			}
		}
	}

	return o.submit(ctx, order, tracker)
}

// submit an order for matching and store it. Returns true if matched (partially or fully), false if not.
func (o *OrderBook) submit(ctx context.Context, order Order, tracker OrderTracker) (bool, error) {
	var matched bool

	if order.IsBid() {
		// order is a bid, match with asks
		matched, _ = o.matchOrder(ctx, tracker.Price, &order, o.orders.Asks)
	} else {
		// order is an ask, match with bids
		matched, _ = o.matchOrder(ctx, tracker.Price, &order, o.orders.Bids)
	}

	addToBooks := true

	if order.Params.Is(ConditionIOC) && !order.IsFilled() {
		order.Cancel()                                             // cancel the rest of the order
		if err := o.orderRepo.SaveOrder(ctx, &order); err != nil { // store the order (not in the books)
			return matched, err
		}
		addToBooks = false // don't add the order to the books (keep it stored but not active)
	}

	if !order.IsFilled() && addToBooks {
		o.addToBooks(tracker)
		if err := o.storeOrder(ctx, order); err != nil {
			return matched, err
		}
	}
	return matched, nil
}

func (o *OrderBook) matchOrder(ctx context.Context, orderPrice float64, order *Order, offers *treemap.TreeMap[OrderTracker, bool]) (matched bool, err error) {
	var (
		buyer, seller          string
		bidOrderID, askOrderID string
	)

	buying := order.Side == SideBuy
	if buying {
		buyer = order.CustomerID
		bidOrderID = order.ID
	} else {
		// sell
		seller = order.CustomerID
		askOrderID = order.ID
	}

	removeOrders := make([]string, 0)
	defer func() {
		for _, orderID := range removeOrders {
			o.removeFromBooks(ctx, orderID)
		}
	}()

	for iter := offers.Iterator(); iter.Valid(); iter.Next() {
		oppositePartialOrder := iter.Key()
		oppositeOrder, ok := o.findActiveOrder(oppositePartialOrder.ID)
		if !ok {
			panic("should NEVER happen - tracker exists but active order does not")
		}

		if oppositeOrder.IsCancelled() {
			removeOrders = append(removeOrders, oppositeOrder.ID) // mark order for removal
			continue                                              // don't match with this order
		}

		qty := min(order.UnfilledQty(), oppositeOrder.UnfilledQty())
		// ensure AONs are complete filled

		// require AON but couldn't fill the order in one trade
		if order.Params.Is(ConditionAON) && qty != order.UnfilledQty() {
			continue
		}

		// other offer requires AON but our order can't fill it completely
		if oppositeOrder.Params.Is(ConditionAON) && qty != oppositeOrder.UnfilledQty() {
			continue
		}

		var price apd.Decimal
		var fPrice float64

		// look only after the best available price
		switch order.Kind {
		case KindMarket:
			switch oppositeOrder.Kind {
			// two opposing market orders are usually forbidden (rejected) - continue matching
			case KindMarket:
				continue
			case KindLimit:
				// crossing the spread
				price = oppositeOrder.Price
				fPrice = oppositePartialOrder.Price
			default:
				// handle error
			}
		case KindLimit:
			myPrice := order.Price
			if buying {
				switch oppositeOrder.Kind {
				case KindMarket:
					price = myPrice
					fPrice = orderPrice
				case KindLimit:
					// check if we can cross the spread
					if myPrice.Cmp(&oppositeOrder.Price) < 0 {
						// other prices are going to be even higher than our limit
						// return false
						return matched, nil
					} else {
						// our bid is higher or equal to their ask - set price to myPrice
						// e.g.: our bid is $20.10, their ask is $20 - trade executes at $20.10
						price = myPrice
						fPrice = orderPrice
					}
				default:
					// handle error
				}
			} else {
				// we're selling
				switch oppositeOrder.Kind {
				// we have a limit, they are buying at our specified price
				case KindMarket:
					price = myPrice
					fPrice = orderPrice
				case KindLimit:
					// check if we can cross the spread
					if myPrice.Cmp(&oppositeOrder.Price) > 0 {
						// we can't match since our ask is higher than the best bid
						return matched, nil
					} else {
						// our ask is lower or equal to their bid - match!
						price = oppositeOrder.Price // set price to their bid
						fPrice = oppositePartialOrder.Price
					}
				default:
					// handle error
				}
			}
		default:
			// handle error
		}

		if buying {
			seller = oppositeOrder.CustomerID
			askOrderID = oppositeOrder.ID
		} else {
			buyer = oppositeOrder.CustomerID
			bidOrderID = oppositeOrder.ID
		}

		order.FilledQty += qty
		oppositeOrder.FilledQty += qty

		matched = true
		// if the other order is filled completely - remove it from the order book
		if oppositeOrder.UnfilledQty() == 0 {
			removeOrders = append(removeOrders, oppositeOrder.ID)
		} else {
			// otherwise update it
			if err := o.updateActiveOrder(ctx, oppositeOrder); err != nil { // otherwise update it
				return matched, err
			}
		}

		event := EventTradeSuccess{
			ID:           uuid.New().String(),
			Buyer:        buyer,
			Seller:       seller,
			TickerSymbol: o.TickerSymbol,
			Qty:          qty,
			Price:        price,
			Total:        apd.Decimal{},
			Timestamp:    time.Now(),
			BidOrderID:   bidOrderID,
			AskOrderID:   askOrderID,
		}

		// fmt.Printf("%#v\n", event)
		o.TradeEvents <- event
		o.SetMarketPrice(ctx, price, fPrice)
		// update tradeBook
		if order.IsFilled() {
			return true, nil
		}
	}

	return matched, nil
}
