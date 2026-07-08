package product

import "github.com/shopspring/decimal"

// Money is a monetary amount backed by an exact decimal, never a float, so
// values survive arithmetic and JSON/database round-trips without the rounding
// error that comes with binary floating point.
//
// It embeds decimal.Decimal, which already satisfies sql.Scanner and
// driver.Valuer, so GORM reads and writes it to the NUMERIC(10,2) column
// directly. Only the JSON encoding is customized: money is always rendered
// with exactly two decimal places (e.g. 1749.00, not 1749).
type Money struct {
	decimal.Decimal
}

// MarshalJSON renders the amount as a JSON number fixed to two decimal places.
func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(m.StringFixed(2)), nil
}

// UnmarshalJSON parses the amount from a JSON number using decimal's exact
// text-based parsing, so no precision is lost decoding the request body.
func (m *Money) UnmarshalJSON(data []byte) error {
	return m.Decimal.UnmarshalJSON(data)
}
