#!/usr/bin/env sh
# Introduce a realistic off-by-one bug and commit it on a new branch, so the
# fix-review-approve workflow has a red test to work on.
#
# The bug: an order whose discounted subtotal is exactly $50.00 is charged
# shipping, although free shipping starts at $50.00. One test catches it:
# TestPrice/order_exactly_at_the_threshold_ships_free.
#
# The workflow builds its workspace from committed code, so the bug must be
# committed, not just saved. Undo with: git checkout main && git branch -D try/break-shipping
set -eu

cd "$(dirname "$0")/.."

if ! git diff --quiet || ! git diff --cached --quiet; then
	echo "break-it: commit or stash your changes first" >&2
	exit 1
fi

file=pricing/pricing.go
from='discounted > 0 \&\& discounted < FreeShippingThreshold'
to='discounted > 0 \&\& discounted <= FreeShippingThreshold'

if ! grep -q 'discounted < FreeShippingThreshold' "$file"; then
	echo "break-it: $file does not contain the line this script edits; already broken?" >&2
	exit 1
fi

git switch -c try/break-shipping
sed "s/$from/$to/" "$file" > "$file.tmp" && mv "$file.tmp" "$file"
git commit -q -am "pricing: charge shipping at the free-shipping threshold (deliberate bug)"

echo
echo "Committed the bug on branch try/break-shipping. The test now fails:"
go test ./... || true
