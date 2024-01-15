package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	type args struct {
		tracker OrderTracker
	}
	tests := []struct {
		name    string
		args    args
		AsksLen int
		BidsLen int
	}{
		{
			name: "SuccessAddAsk",
			args: args{
				tracker: OrderTracker{
					ID:        "1",
					Kind:      KindMarket,
					Price:     10,
					Side:      SideSell,
					Timestamp: time.Now().UnixNano(),
				},
			},
			AsksLen: 1,
		},
		{
			name: "SuccessAddBids",
			args: args{
				tracker: OrderTracker{
					ID:        "1",
					Kind:      KindMarket,
					Price:     10,
					Side:      SideBuy,
					Timestamp: time.Now().UnixNano(),
				},
			},
			BidsLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// bid need price high
			bid := newComparator(true)
			// asker need price lower
			ask := newComparator(false)
			set := NewOrderSet(bid, ask)
			set.Add(tt.args.tracker)

			assert.Equal(t, tt.AsksLen, set.Asks.Len())
			assert.Equal(t, tt.BidsLen, set.Bids.Len())
			assert.Equal(t, tt.AsksLen+tt.BidsLen, len(set.OrderTrackers))
		})
	}
}
