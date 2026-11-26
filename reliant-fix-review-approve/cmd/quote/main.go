// Command quote prices a cart from the command line.
//
//	go run ./cmd/quote -code WELCOME10 kettle:5000:1 mug:1200:2
//
// Each argument is sku:unit_cents:quantity.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"example.com/checkout/pricing"
)

func main() {
	code := flag.String("code", "", "discount code")
	flag.Parse()

	var items []pricing.Item
	for _, arg := range flag.Args() {
		parts := strings.Split(arg, ":")
		if len(parts) != 3 {
			fail("want sku:unit_cents:quantity, got %q", arg)
		}
		unit, err1 := strconv.ParseInt(parts[1], 10, 64)
		qty, err2 := strconv.ParseInt(parts[2], 10, 64)
		if err1 != nil || err2 != nil {
			fail("bad number in %q", arg)
		}
		items = append(items, pricing.Item{SKU: parts[0], UnitCents: unit, Quantity: qty})
	}

	q, err := pricing.Price(items, *code)
	if err != nil {
		fail("%v", err)
	}
	fmt.Printf("subtotal %s\ndiscount -%s\nshipping %s\ntax      %s\ntotal    %s\n",
		money(q.Subtotal), money(q.Discount), money(q.Shipping), money(q.Tax), money(q.Total))
}

func money(cents int64) string { return fmt.Sprintf("$%d.%02d", cents/100, cents%100) }

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "quote: "+format+"\n", args...)
	os.Exit(2)
}
