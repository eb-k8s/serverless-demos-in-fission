package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type PlaceOrderRequest struct {
	UserId       string          `json:"user_id,omitempty"`
	UserCurrency string          `json:"user_currency,omitempty"`
	Address      *Address        `json:"address,omitempty"`
	Email        string          `json:"email,omitempty"`
	CreditCard   *CreditCardInfo `json:"credit_card,omitempty"`
}

type Address struct {
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	Country       string `json:"country,omitempty"`
	ZipCode       int32  `json:"zip_code,omitempty"`
}

type CreditCardInfo struct {
	CreditCardNumber          string `json:"credit_card_number,omitempty"`
	CreditCardCvv             int32  `json:"credit_card_cvv,omitempty"`
	CreditCardExpirationYear  int32  `json:"credit_card_expiration_year,omitempty"`
	CreditCardExpirationMonth int32  `json:"credit_card_expiration_month,omitempty"`
}

type PlaceOrderResponse struct {
	Order *OrderResult `json:"order,omitempty"`
}

func (m *PlaceOrderResponse) GetOrder() *OrderResult {
	if m != nil {
		return m.Order
	}
	return nil
}

type OrderResult struct {
	OrderId            string       `json:"order_id,omitempty"`
	ShippingTrackingId string       `json:"shipping_tracking_id,omitempty"`
	ShippingCost       *Money       `json:"shipping_cost,omitempty"`
	ShippingAddress    *Address     `json:"shipping_address,omitempty"`
	Items              []*OrderItem `json:"items,omitempty"`
}

func (m *OrderResult) GetOrderId() string {
	if m != nil {
		return m.OrderId
	}
	return ""
}

func (m *OrderResult) GetShippingTrackingId() string {
	if m != nil {
		return m.ShippingTrackingId
	}
	return ""
}

func (m *OrderResult) GetShippingCost() *Money {
	if m != nil {
		return m.ShippingCost
	}
	return nil
}

func (m *OrderResult) GetShippingAddress() *Address {
	if m != nil {
		return m.ShippingAddress
	}
	return nil
}

func (m *OrderResult) GetItems() []*OrderItem {
	if m != nil {
		return m.Items
	}
	return nil
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

type OrderItem struct {
	Item *CartItem `json:"item,omitempty"`
	Cost *Money    `json:"cost,omitempty"`
}

func (m *OrderItem) GetItem() *CartItem {
	if m != nil {
		return m.Item
	}
	return nil
}

func (m *OrderItem) GetCost() *Money {
	if m != nil {
		return m.Cost
	}
	return nil
}

type CartItem struct {
	ProductId string `json:"product_id,omitempty"`
	Quantity  int32  `json:"quantity,omitempty"`
}

func (m *CartItem) GetProductId() string {
	if m != nil {
		return m.ProductId
	}
	return ""
}

func (m *CartItem) GetQuantity() int32 {
	if m != nil {
		return m.Quantity
	}
	return 0
}

type GetCartRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type Cart struct {
	UserId string      `json:"user_id,omitempty"`
	Items  []*CartItem `json:"items,omitempty"`
}

func (m *Cart) GetItems() []*CartItem {
	if m != nil {
		return m.Items
	}
	return nil
}

func GetCart(ctx context.Context, client http.Client, cartSvcAddr string, in *GetCartRequest) (*Cart, error) {
	out := new(Cart)
	v := url.Values{}
	v.Add("user_id", in.UserId)
	req, err := http.NewRequestWithContext(ctx, "GET", cartSvcAddr+"?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type EmptyCartRequest struct {
	UserId string `json:"user_id,omitempty"`
}

func EmptyCart(ctx context.Context, client http.Client, cartSvcAddr string, in *EmptyCartRequest) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", cartSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

type GetProductRequest struct {
	Id string `json:"id,omitempty"`
}

type Product struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Picture     string `json:"picture,omitempty"`
	PriceUsd    *Money `json:"price_usd,omitempty"`
	// Categories such as "clothing" or "kitchen" that can be used to look up
	// other related products.
	Categories []string `json:"categories,omitempty"`
}

func (m *Product) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Product) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Product) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Product) GetPicture() string {
	if m != nil {
		return m.Picture
	}
	return ""
}

func (m *Product) GetPriceUsd() *Money {
	if m != nil {
		return m.PriceUsd
	}
	return nil
}

func (m *Product) GetCategories() []string {
	if m != nil {
		return m.Categories
	}
	return nil
}

func GetProduct(ctx context.Context, client http.Client, productCatalogSvcAddr string, in *GetProductRequest) (*Product, error) {
	out := new(Product)
	v := url.Values{}
	v.Add("id", in.Id)
	req, err := http.NewRequestWithContext(ctx, "GET", productCatalogSvcAddr+"?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type CurrencyConversionRequest struct {
	From *Money `json:"from,omitempty"`
	// The 3-letter currency code defined in ISO 4217.
	ToCode string `json:"to_code,omitempty"`
}

func Convert(ctx context.Context, client http.Client, currencySvcAddr string, in *CurrencyConversionRequest) (*Money, error) {
	out := new(Money)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", currencySvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ChargeRequest struct {
	Amount     *Money          `json:"amount,omitempty"`
	CreditCard *CreditCardInfo `json:"credit_card,omitempty"`
}

func (m *ChargeRequest) GetAmount() *Money {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *ChargeRequest) GetCreditCard() *CreditCardInfo {
	if m != nil {
		return m.CreditCard
	}
	return nil
}

type ChargeResponse struct {
	TransactionId string `json:"transaction_id,omitempty"`
}

func (m *ChargeResponse) GetTransactionId() string {
	if m != nil {
		return m.TransactionId
	}
	return ""
}

func Charge(ctx context.Context, client http.Client, paymentSvcAddr string, in *ChargeRequest) (*ChargeResponse, error) {
	out := new(ChargeResponse)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", paymentSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type SendOrderConfirmationRequest struct {
	Email string       `json:"email,omitempty"`
	Order *OrderResult `json:"order,omitempty"`
}

func (m *SendOrderConfirmationRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *SendOrderConfirmationRequest) GetOrder() *OrderResult {
	if m != nil {
		return m.Order
	}
	return nil
}

func SendOrderConfirmation(ctx context.Context, client http.Client, emailSvcAddr string, in *SendOrderConfirmationRequest) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", emailSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

type ShipOrderRequest struct {
	Address *Address    `json:"address,omitempty"`
	Items   []*CartItem `json:"items,omitempty"`
}

func (m *ShipOrderRequest) GetAddress() *Address {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *ShipOrderRequest) GetItems() []*CartItem {
	if m != nil {
		return m.Items
	}
	return nil
}

type ShipOrderResponse struct {
	TrackingId string `json:"tracking_id,omitempty"`
}

func (m *ShipOrderResponse) GetTrackingId() string {
	if m != nil {
		return m.TrackingId
	}
	return ""
}

func ShipOrder(ctx context.Context, client http.Client, shippingSvcAddr string, in *ShipOrderRequest) (*ShipOrderResponse, error) {
	out := new(ShipOrderResponse)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", shippingSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type GetQuoteRequest struct {
	Address *Address    ` json:"address,omitempty"`
	Items   []*CartItem ` json:"items,omitempty"`
}

func (m *GetQuoteRequest) GetAddress() *Address {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *GetQuoteRequest) GetItems() []*CartItem {
	if m != nil {
		return m.Items
	}
	return nil
}

type GetQuoteResponse struct {
	CostUsd *Money `json:"cost_usd,omitempty"`
}

func (m *GetQuoteResponse) GetCostUsd() *Money {
	if m != nil {
		return m.CostUsd
	}
	return nil
}

func GetQuote(ctx context.Context, client http.Client, shippingSvcAddr string, in *GetQuoteRequest) (*GetQuoteResponse, error) {
	out := new(GetQuoteResponse)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", shippingSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
