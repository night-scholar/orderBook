package main

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
)

// Order strores information about request
type Order struct {
	side      Side
	id        string
	trader    string
 	timestamp string
	quantity  decimal.Decimal
	price     decimal.Decimal
	data	  string
	signature string
	perpetual string
	broker	  string
}

// NewOrder creates new constant object Order
func NewOrder(orderID string, side Side, trader string,quantity, price decimal.Decimal, timestamp ,data,signature,perpetual,broker string) *Order {
	return &Order{
		id:        orderID,
		side:      side,
		trader:    trader,
		quantity:  quantity,
		price:     price,
		timestamp: timestamp,
		data: 	   data,
		signature: signature,
		perpetual: perpetual,
		broker:	   broker,
	}
}

// ID returns orderID field copy
func (o *Order) ID() string {
	return o.id
}

// Side returns side of the order
func (o *Order) Side() Side {
	return o.side
}

func (o *Order) Trader() string {
	return o.trader
}

// Quantity returns quantity field copy
func (o *Order) Quantity() decimal.Decimal {
	return o.quantity
}

// Price returns price field copy
func (o *Order) Price() decimal.Decimal {
	return o.price
}

// Time returns timestamp field copy
func (o *Order) Time() string {
	return o.timestamp
}

// String implements Stringer interface
func (o *Order) String() string {
	return fmt.Sprintf("\n\"%s\":\n\tside: %s\n\tquantity: %s\n\tprice: %s\n\ttime: %s\n", o.ID(), o.Side(), o.Quantity(), o.Price(), o.Time())
}

// MarshalJSON implements json.Marshaler interface
func (o *Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		&struct {
			S         Side            `json:"side"`
			ID        string          `json:"id"`
			Trader    string          `json:"trader"`
			Timestamp string       `json:"timestamp"`
			Quantity  decimal.Decimal `json:"quantity"`
			Price     decimal.Decimal `json:"price"`
		}{
			S:         o.Side(),
			ID:        o.ID(),
			Trader:    o.Trader(),
			Timestamp: o.Time(),
			Quantity:  o.Quantity(),
			Price:     o.Price(),
		},
	)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (o *Order) UnmarshalJSON(data []byte) error {
	obj := struct {
		S         Side            `json:"side"`
		ID        string          `json:"id"`
		Trader    string          `json:"trader`
		Timestamp string      `json:"timestamp"`
		Quantity  decimal.Decimal `json:"quantity"`
		Price     decimal.Decimal `json:"price"`
	}{}

	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	o.side = obj.S
	o.id = obj.ID
	o.trader = obj.Trader
	o.timestamp = obj.Timestamp
	o.quantity = obj.Quantity
	o.price = obj.Price
	return nil
}
