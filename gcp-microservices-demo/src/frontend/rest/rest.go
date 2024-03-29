package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type GetSupportedCurrenciesResponse struct {
	CurrencyCodes []string `json:"currency_codes,omitempty"`
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

type ListProductsResponse struct {
	Products []*Product `json:"products,omitempty"`
}

type GetProductRequest struct {
	Id string `json:"id,omitempty"`
}

type GetCartRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type Cart struct {
	UserId string      `json:"user_id,omitempty"`
	Items  []*CartItem `json:"items,omitempty"`
}

type CartItem struct {
	ProductId string `json:"product_id,omitempty"`
	Quantity  int32  `json:"quantity,omitempty"`
}

type EmptyCartRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type AddItemRequest struct {
	UserId string    `json:"user_id,omitempty"`
	Item   *CartItem `json:"item,omitempty"`
}

type CurrencyConversionRequest struct {
	From *Money `json:"from,omitempty"`
	// The 3-letter currency code defined in ISO 4217.
	ToCode string `json:"to_code,omitempty"`
}

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

type GetQuoteResponse struct {
	CostUsd *Money `json:"cost_usd,omitempty"`
}

type ListRecommendationsRequest struct {
	UserId     string   `json:"user_id,omitempty"`
	ProductIds []string `json:"product_ids,omitempty"`
}

type ListRecommendationsResponse struct {
	ProductIds []string `json:"product_ids,omitempty"`
}

type Ad struct {
	// url to redirect to when an ad is clicked.
	RedirectUrl string `json:"redirect_url,omitempty"`
	// short advertisement text to display.
	Text string `json:"text,omitempty"`
}

type AdRequest struct {
	// List of important key words from the current page describing the context.
	ContextKeys []string `json:"context_keys,omitempty"`
}

type AdResponse struct {
	Ads []*Ad `json:"ads,omitempty"`
}

type CreditCardInfo struct {
	CreditCardNumber          string `json:"credit_card_number,omitempty"`
	CreditCardCvv             int32  `json:"credit_card_cvv,omitempty"`
	CreditCardExpirationYear  int32  `json:"credit_card_expiration_year,omitempty"`
	CreditCardExpirationMonth int32  `json:"credit_card_expiration_month,omitempty"`
}

type PlaceOrderRequest struct {
	UserId       string          `json:"user_id,omitempty"`
	UserCurrency string          `json:"user_currency,omitempty"`
	Address      *Address        `json:"address,omitempty"`
	Email        string          `json:"email,omitempty"`
	CreditCard   *CreditCardInfo `json:"credit_card,omitempty"`
}

type OrderResult struct {
	OrderId            string       `json:"order_id,omitempty"`
	ShippingTrackingId string       `json:"shipping_tracking_id,omitempty"`
	ShippingCost       *Money       `json:"shipping_cost,omitempty"`
	ShippingAddress    *Address     `pjson:"shipping_address,omitempty"`
	Items              []*OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	Item *CartItem `json:"item,omitempty"`
	Cost *Money    `json:"cost,omitempty"`
}

type PlaceOrderResponse struct {
	Order *OrderResult `json:"order,omitempty"`
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

func GetSupportedCurrencies(currencySvcAddr string) (*GetSupportedCurrenciesResponse, error) {
	out := new(GetSupportedCurrenciesResponse)
	res, err := http.Get(currencySvcAddr)
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

func ListProducts(productCatalogSvcAddr string) (*ListProductsResponse, error) {
	out := new(ListProductsResponse)
	res, err := http.Get(productCatalogSvcAddr)
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

func (m *ListProductsResponse) GetProducts() []*Product {
	if m != nil {
		return m.Products
	}
	return nil
}

func GetProduct(productCatalogSvcAddr string, in *GetProductRequest) (*Product, error) {
	out := new(Product)
	v := url.Values{}
	v.Add("id", in.Id)
	res, err := http.Get(productCatalogSvcAddr + "?" + v.Encode())
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

func GetCart(cartSvcAddr string, in *GetCartRequest) (*Cart, error) {
	out := new(Cart)
	v := url.Values{}
	v.Add("user_id", in.UserId)
	res, err := http.Get(cartSvcAddr + "?" + v.Encode())
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

func (m *Cart) GetItems() []*CartItem {
	if m != nil {
		return m.Items
	}
	return nil
}

func EmptyCart(cartSvcAddr string, in *EmptyCartRequest) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("DELETE", cartSvcAddr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func AddItem(cartSvcAddr string, in *AddItemRequest) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	_, err = http.Post(cartSvcAddr, "application/json", bytes.NewBuffer(payload))
	return err
}

func Convert(currencySvcAddr string, in *CurrencyConversionRequest) (*Money, error) {
	out := new(Money)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(currencySvcAddr, "application/json", bytes.NewBuffer(payload))
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

func GetQuote(shippingSvcAddr string, in *GetQuoteRequest) (*GetQuoteResponse, error) {
	out := new(GetQuoteResponse)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(shippingSvcAddr, "application/json", bytes.NewBuffer(payload))
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

func (m *GetQuoteResponse) GetCostUsd() *Money {
	if m != nil {
		return m.CostUsd
	}
	return nil
}

func (m *ListRecommendationsResponse) GetProductIds() []string {
	if m != nil {
		return m.ProductIds
	}
	return nil
}

func ListRecommendations(recommendationSvcAddr string, in *ListRecommendationsRequest) (*ListRecommendationsResponse, error) {
	out := new(ListRecommendationsResponse)
	v := url.Values{}
	v.Add("user_id", in.UserId)
	v.Add("product_ids", strings.Join(in.ProductIds, ","))
	res, err := http.Get(recommendationSvcAddr + "?" + v.Encode())
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

func GetAds(adSvcAddr string, in *AdRequest) (*AdResponse, error) {
	out := new(AdResponse)
	v := url.Values{}
	v.Add("context_keys", strings.Join(in.ContextKeys, ","))
	res, err := http.Get(adSvcAddr + "?" + v.Encode())
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

func PlaceOrder(checkoutSvcAddr string, in *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	out := new(PlaceOrderResponse)
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(checkoutSvcAddr, "application/json", bytes.NewBuffer(payload))
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

func (m *AdResponse) GetAds() []*Ad {
	if m != nil {
		return m.Ads
	}
	return nil
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

func (m *PlaceOrderResponse) GetOrder() *OrderResult {
	if m != nil {
		return m.Order
	}
	return nil
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
