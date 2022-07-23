package money

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// Scan implements sql.Scanner to deserialize a Money instance from a "currency_code amount" space-separated pair or an INTEGER column
// (with the currency already defaulted)
func (m *Money) Scan(src interface{}) error {
	amount := &Amount{}
	currency := &Currency{}

	// let's support string and int64
	switch x := src.(type) {
	// currency_code amount
	case string:
		parts := strings.Split(x, " ")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("%#v is not valid to scan into Money", x)
		}
		if err := currency.Scan(parts[0]); err != nil {
			return fmt.Errorf("scanning %#v into a Currency: %v", parts[0], err)
		}
		if err := amount.Scan(parts[1]); err != nil {
			return fmt.Errorf("scanning %#v into an Amount: %v", parts[1], err)
		}
	case int32, int64:
		if err := amount.Scan(x); err != nil {
			return fmt.Errorf("scanning %#v into an Amount: %v", x, err)
		}
	default:
		return fmt.Errorf("don't know how to scan %T into Money", src)
	}

	// allocate new Money with the scanned amount and currency
	*m = Money{
		amount:   amount,
		currency: currency,
	}

	return nil
}

// Value implements driver.Valuer to serialize an Amount into sql/driver compatible type
func (a Amount) Value() (driver.Value, error) {
	return a.val, nil
}

// Scan implements sql.Scanner to deserialize an Amount from a string or integer value read from a database
func (a *Amount) Scan(src interface{}) error {
	var aa Amount
	// let's support string and int64
	switch x := src.(type) {
	case string:
		if i, err := strconv.ParseInt(x, 10, 64); err != nil {
			return fmt.Errorf("failed to parse %#v as int: %v", src, err)
		} else {
			aa.val = i
		}
	case int64:
		aa.val = x
	case int32:
		aa.val = int64(x) // safe upcast
	case int:
		aa.val = int64(x) // safe upcast
	default:
		return fmt.Errorf("%T is not a supported type for an Amount", src)
	}

	// copy the value
	*a = aa

	return nil
}

// Value implements driver.Valuer to serialize a Currency code into a string for saving to a database
func (c Currency) Value() (driver.Value, error) {
	return c.Code, nil
}

// Scan implements sql.Scanner to deserialize a Currency from a string value read from a database
func (c *Currency) Scan(src interface{}) error {
	var val *Currency
	// let's support string only
	switch x := src.(type) {
	case string:
		val = GetCurrency(x)
	default:
		return fmt.Errorf("%T is not a supported type for a Currency (store the Currency.Code value as a string only)", src)
	}

	if val == nil {
		return fmt.Errorf("getCurrency(%#v) returned nil", src)
	}

	// copy the value
	*c = *val

	return nil
}
