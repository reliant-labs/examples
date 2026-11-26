// Package pricing turns a cart into an order total: subtotal, discount code,
// shipping and sales tax. All money is in integer cents.
package pricing

import (
	"errors"
	"fmt"
	"strings"
)

// FreeShippingThreshold is the discounted subtotal, in cents, at which
// shipping becomes free.
const FreeShippingThreshold int64 = 50_00

// FlatShipping is charged on orders below FreeShippingThreshold.
const FlatShipping int64 = 5_99

// TaxRateBasisPoints is the sales tax rate in basis points (825 = 8.25%).
const TaxRateBasisPoints int64 = 825

// ErrUnknownCode is returned for a discount code that does not exist.
var ErrUnknownCode = errors.New("unknown discount code")

// Item is one line of a cart.
type Item struct {
	SKU       string
	UnitCents int64
	Quantity  int64
}

// Quote is the price breakdown for a cart.
type Quote struct {
	Subtotal int64
	Discount int64
	Shipping int64
	Tax      int64
	Total    int64
}

// percentOff maps a discount code to the percentage it takes off the subtotal.
var percentOff = map[string]int64{
	"WELCOME10": 10,
	"HALFOFF":   50,
}

// Price computes the quote for items with an optional discount code.
// Codes are case-insensitive; an empty code means no discount.
func Price(items []Item, code string) (Quote, error) {
	var q Quote
	for _, it := range items {
		if it.Quantity < 0 || it.UnitCents < 0 {
			return Quote{}, fmt.Errorf("item %q: negative quantity or price", it.SKU)
		}
		q.Subtotal += it.UnitCents * it.Quantity
	}

	if code != "" {
		pct, ok := percentOff[strings.ToUpper(strings.TrimSpace(code))]
		if !ok {
			return Quote{}, fmt.Errorf("%w: %q", ErrUnknownCode, code)
		}
		q.Discount = q.Subtotal * pct / 100
	}

	discounted := q.Subtotal - q.Discount
	if discounted > 0 && discounted < FreeShippingThreshold {
		q.Shipping = FlatShipping
	}

	q.Tax = roundHalfUp(discounted*TaxRateBasisPoints, 10_000)
	q.Total = discounted + q.Shipping + q.Tax
	return q, nil
}

// roundHalfUp divides n by d (both non-negative) and rounds half up.
func roundHalfUp(n, d int64) int64 {
	return (n + d/2) / d
}
