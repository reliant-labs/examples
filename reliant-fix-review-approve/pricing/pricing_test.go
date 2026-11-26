package pricing

import (
	"errors"
	"testing"
)

func TestPrice(t *testing.T) {
	tests := []struct {
		name  string
		items []Item
		code  string
		want  Quote
	}{
		{
			name:  "empty cart costs nothing",
			items: nil,
			want:  Quote{},
		},
		{
			name:  "small order pays flat shipping",
			items: []Item{{SKU: "mug", UnitCents: 12_00, Quantity: 2}},
			// 24.00 + 5.99 shipping + 1.98 tax (8.25% of 24.00)
			want: Quote{Subtotal: 24_00, Shipping: 5_99, Tax: 1_98, Total: 31_97},
		},
		{
			// The one to break: see README "Break a test".
			name:  "order exactly at the threshold ships free",
			items: []Item{{SKU: "kettle", UnitCents: 50_00, Quantity: 1}},
			// 50.00 + 0 shipping + 4.13 tax (4.125 rounds half up)
			want: Quote{Subtotal: 50_00, Shipping: 0, Tax: 4_13, Total: 54_13},
		},
		{
			name:  "discount is applied before the shipping threshold",
			items: []Item{{SKU: "kettle", UnitCents: 50_00, Quantity: 1}},
			code:  "welcome10",
			// 50.00 - 5.00 = 45.00, below 50.00, so shipping is charged
			want: Quote{Subtotal: 50_00, Discount: 5_00, Shipping: 5_99, Tax: 3_71, Total: 54_70},
		},
		{
			name: "multiple lines add up",
			items: []Item{
				{SKU: "beans", UnitCents: 18_50, Quantity: 3},
				{SKU: "filter", UnitCents: 4_25, Quantity: 2},
			},
			code: "HALFOFF",
			// 55.50 + 8.50 = 64.00, half off = 32.00, + 5.99, + 2.64 tax
			want: Quote{Subtotal: 64_00, Discount: 32_00, Shipping: 5_99, Tax: 2_64, Total: 40_63},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Price(tt.items, tt.code)
			if err != nil {
				t.Fatalf("Price() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Price() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPriceRejectsUnknownCode(t *testing.T) {
	_, err := Price([]Item{{SKU: "mug", UnitCents: 12_00, Quantity: 1}}, "FREESTUFF")
	if !errors.Is(err, ErrUnknownCode) {
		t.Fatalf("Price() error = %v, want ErrUnknownCode", err)
	}
}

func TestPriceRejectsNegativeQuantity(t *testing.T) {
	_, err := Price([]Item{{SKU: "mug", UnitCents: 12_00, Quantity: -1}}, "")
	if err == nil {
		t.Fatal("Price() error = nil, want an error for a negative quantity")
	}
}

func TestRoundHalfUp(t *testing.T) {
	for _, tt := range []struct{ n, d, want int64 }{
		{41_250, 10_000, 4},
		{45_000, 10_000, 5},
		{44_999, 10_000, 4},
		{0, 10_000, 0},
	} {
		if got := roundHalfUp(tt.n, tt.d); got != tt.want {
			t.Errorf("roundHalfUp(%d, %d) = %d, want %d", tt.n, tt.d, got, tt.want)
		}
	}
}
