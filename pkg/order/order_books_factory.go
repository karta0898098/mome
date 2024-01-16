package order

import "github.com/cockroachdb/apd"

type BooksFactory struct {
	Mode      string
	OrderRepo Repository
}

func (config *BooksFactory) Create() map[string]*OrderBook {
	switch config.Mode {
	case "demo":
		return newExampleOrderBooks(config.OrderRepo)
	case "fetch":
		panic("not yet implement")
	}

	return nil
}

func newExampleOrderBooks(repo Repository) map[string]*OrderBook {
	orderBooks := map[string]*OrderBook{
		"TEST": NewOrderBook("TEST", *apd.New(2025, -2), repo),
	}

	return orderBooks
}
