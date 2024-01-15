package order

import "testing"

func TestSet_Add(t *testing.T) {
	type fields struct {
		Bids          *treemap.TreeMap[OrderTracker, bool]
		Asks          *treemap.TreeMap[OrderTracker, bool]
		PartialOrders map[string]OrderTracker
	}
	type args struct {
		partialOrder OrderTracker
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				Bids:          tt.fields.Bids,
				Asks:          tt.fields.Asks,
				PartialOrders: tt.fields.PartialOrders,
			}
			set.Add(tt.args.partialOrder)
		})
	}
}
