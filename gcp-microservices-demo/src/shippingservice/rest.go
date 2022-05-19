package main

type GetQuoteRequest struct {
	Address *Address    ` json:"address,omitempty"`
	Items   []*CartItem ` json:"items,omitempty"`
}

type Address struct {
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	Country       string `json:"country,omitempty"`
	ZipCode       int32  `json:"zip_code,omitempty"`
}

type CartItem struct {
	ProductId string `json:"product_id,omitempty"`
	Quantity  int32  `json:"quantity,omitempty"`
}

type GetQuoteResponse struct {
	CostUsd *Money `json:"cost_usd,omitempty"`
}

// Represents an amount of money with its currency type.
type Money struct {
	// The 3-letter currency code defined in ISO 4217.
	CurrencyCode string `json:"currency_code,omitempty"`
	// The whole units of the amount.
	// For example if `currencyCode` is `"USD"`, then 1 unit is one US dollar.
	Units int64 `json:"units,omitempty"`
	// Number of nano (10^-9) units of the amount.
	// The value must be between -999,999,999 and +999,999,999 inclusive.
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int32 `json:"nanos,omitempty"`
}

type ShipOrderRequest struct {
	Address *Address    `json:"address,omitempty"`
	Items   []*CartItem `json:"items,omitempty"`
}

type ShipOrderResponse struct {
	TrackingId string `json:"tracking_id,omitempty"`
}

func (m *Money) GetCurrencyCode() string {
	if m != nil {
		return m.CurrencyCode
	}
	return ""
}

func (m *Money) GetUnits() int64 {
	if m != nil {
		return m.Units
	}
	return 0
}

func (m *Money) GetNanos() int32 {
	if m != nil {
		return m.Nanos
	}
	return 0
}
