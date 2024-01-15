package order

import (
	"sort"

	"github.com/igrmk/treemap/v2"
)

type Set struct {
	Bids *treemap.TreeMap[OrderTracker, bool]
	Asks *treemap.TreeMap[OrderTracker, bool]

	OrderTrackers map[string]OrderTracker
}

type Comparator func(x, y OrderTracker) bool

func NewOrderSet(bidComparator Comparator, askComparator Comparator) *Set {
	return &Set{
		Bids:          treemap.NewWithKeyCompare[OrderTracker, bool](bidComparator),
		Asks:          treemap.NewWithKeyCompare[OrderTracker, bool](askComparator),
		OrderTrackers: make(map[string]OrderTracker),
	}
}

// FIFO
// ref : https://corporatefinanceinstitute.com/resources/career-map/sell-side/capital-markets/matching-orders/
func newComparator(priceDescending bool) Comparator {
	const (
		ascending  bool = true
		descending bool = false
	)

	sort := ascending
	if priceDescending {
		sort = descending
	}

	return func(x, y OrderTracker) bool {
		// market orders first
		if x.Kind == KindMarket && y.Kind != KindMarket {
			return true
		} else if x.Kind != KindMarket && y.Kind == KindMarket {
			return false
		} else if x.Kind == KindMarket && y.Kind == KindMarket {
			// both market orders sort by time
			return x.Timestamp < y.Timestamp
		}

		diff := x.Price - y.Price
		// if prices are equal, compare timestamps
		if diff == 0 {
			return x.Timestamp < y.Timestamp
		}

		// if x price is less than y return true if ascending, false if descending
		if diff < 0 {
			return sort
		}

		// if x price is bigger than y return false if ascending, true if descending
		return !sort
	}
}

func newStopComparator(priceDescending bool) Comparator {
	const (
		ascending  bool = true
		descending bool = false
	)
	sort := ascending
	if priceDescending {
		sort = descending
	}
	return func(x, y OrderTracker) bool { // ignores order types because we're always comparing stop prices
		priceCmp := x.Price - y.Price // compare prices
		if priceCmp == 0 {            // if prices are equal, compare timestamps
			return x.Timestamp < y.Timestamp
		}
		if priceCmp < 0 { // if a price is less than b return true if ascending, false if descending
			return sort
		}
		return !sort // if a price is bigger than b return false if ascending, true if descending
	}
}

func (set *Set) Add(tracker OrderTracker) {
	if tracker.Side == SideBuy {
		set.Bids.Set(tracker, true)
	} else {
		set.Asks.Set(tracker, true)
	}
	set.OrderTrackers[tracker.ID] = tracker
}

func (set *Set) Remove(id string) {
	p, ok := set.OrderTrackers[id]
	if !ok {
		return
	}
	delete(set.OrderTrackers, id)
	if p.Side == SideBuy {
		set.Bids.Del(p)
	} else if p.Side == SideSell {
		set.Asks.Del(p)
	}
}

func (set *Set) Find(id string) (OrderTracker, bool) {
	o, ok := set.OrderTrackers[id]
	return o, ok
}

// Iterator which iterates through sorted bids or asks.
func (set *Set) Iterator(side Side) treemap.ForwardIterator[OrderTracker, bool] {
	if side == SideBuy {
		return set.Bids.Iterator()
	}
	return set.Asks.Iterator()
}

// Len returns the number of bids or asks in the set.
func (set *Set) Len(side Side) int {
	if side == SideBuy {
		return set.Bids.Len()
	}
	return set.Asks.Len()
}

// FindAllAsksAbove ask orders below or equal the price, sorted by time ast
func (set *Set) FindAllAsksAbove(price float64) []OrderTracker {
	results := make([]OrderTracker, 0)

	for iter := set.Asks.Iterator(); iter.Valid(); iter.Next() {
		if iter.Key().Price >= price {
			results = append(results, iter.Key())
		} else {
			// iterator returns a sorted array, if price is bigger we don't have to look any further
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp < results[j].Timestamp
	})

	return results
}

// FindAllBidsBelow bids orders above or equal the price, sorted by time ast
func (set *Set) FindAllBidsBelow(price float64) []OrderTracker {
	results := make([]OrderTracker, 0)

	for iter := set.Bids.Iterator(); iter.Valid(); iter.Next() {
		if iter.Key().Price <= price {
			results = append(results, iter.Key())
		} else {
			// iterator returns a sorted array, if price is bigger we don't have to look any further
			break
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp < results[j].Timestamp
	})

	return results
}
