package model

import "testing"

func TestCanTransitionTo(t *testing.T) {
	cases := []struct {
		from, to OrderStatus
		want     bool
	}{
		{StatusPending, StatusPaid, true},
		{StatusPending, StatusCancelled, true},
		{StatusPending, StatusRefunded, false},
		{StatusPaid, StatusShipped, true},
		{StatusPaid, StatusRefunded, true},
		{StatusPaid, StatusCancelled, false},
		{StatusShipped, StatusRefunded, true},
		{StatusRefunded, StatusPaid, false},
		{StatusCancelled, StatusPaid, false},
	}

	for _, c := range cases {
		o := Order{Status: c.from}
		if got := o.CanTransitionTo(c.to); got != c.want {
			t.Errorf("%s -> %s = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestOrderItemSubtotal(t *testing.T) {
	item := OrderItem{Quantity: 3, UnitPrice: 1999}
	if got := item.Subtotal(); got != 5997 {
		t.Errorf("Subtotal = %d, want 5997", got)
	}
}
