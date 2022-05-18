package main

import (
	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/frontend/rest"
	"github.com/pkg/errors"
)

func (fe *frontendServer) getCurrencies() ([]string, error) {
	currs, err := rest.GetSupportedCurrencies(fe.currencySvcAddr)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, c := range currs.CurrencyCodes {
		if _, ok := whitelistedCurrencies[c]; ok {
			out = append(out, c)
		}
	}
	return out, nil
}

func (fe *frontendServer) getProducts() ([]*rest.Product, error) {
	resp, err := rest.ListProducts(fe.productCatalogSvcAddr)
	return resp.GetProducts(), err
}

func (fe *frontendServer) getProduct(id string) (*rest.Product, error) {
	resp, err := rest.GetProduct(fe.productCatalogSvcAddr, &rest.GetProductRequest{Id: id})
	return resp, err
}

func (fe *frontendServer) getCart(userID string) ([]*rest.CartItem, error) {
	resp, err := rest.GetCart(fe.cartSvcAddr, &rest.GetCartRequest{UserId: userID})
	return resp.GetItems(), err
}

func (fe *frontendServer) emptyCart(userID string) error {
	return rest.EmptyCart(fe.cartSvcAddr, &rest.EmptyCartRequest{UserId: userID})
}

func (fe *frontendServer) insertCart(userID, productID string, quantity int32) error {
	return rest.AddItem(fe.cartSvcAddr, &rest.AddItemRequest{
		UserId: userID,
		Item: &rest.CartItem{
			ProductId: productID,
			Quantity:  quantity},
	})
}

func (fe *frontendServer) convertCurrency(money *rest.Money, currency string) (*rest.Money, error) {
	return rest.Convert(fe.currencySvcAddr, &rest.CurrencyConversionRequest{
		From:   money,
		ToCode: currency})
}

func (fe *frontendServer) getShippingQuote(items []*rest.CartItem, currency string) (*rest.Money, error) {
	quote, err := rest.GetQuote(fe.shippingSvcAddr, &rest.GetQuoteRequest{
		Address: nil,
		Items:   items})
	if err != nil {
		return nil, err
	}
	localized, err := fe.convertCurrency(quote.GetCostUsd(), currency)
	return localized, errors.Wrap(err, "failed to convert currency for shipping cost")
}

func (fe *frontendServer) getRecommendations(userID string, productIDs []string) ([]*rest.Product, error) {
	resp, err := rest.ListRecommendations(fe.recommendationSvcAddr, &rest.ListRecommendationsRequest{UserId: userID, ProductIds: productIDs})
	if err != nil {
		return nil, err
	}
	out := make([]*rest.Product, len(resp.GetProductIds()))
	for i, v := range resp.GetProductIds() {
		p, err := fe.getProduct(v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get recommended product info (#%s)", v)
		}
		out[i] = p
	}
	if len(out) > 4 {
		out = out[:4] // take only first four to fit the UI
	}
	return out, err
}

func (fe *frontendServer) getAd(ctxKeys []string) ([]*rest.Ad, error) {
	resp, err := rest.GetAds(fe.adSvcAddr, &rest.AdRequest{
		ContextKeys: ctxKeys,
	})
	return resp.GetAds(), errors.Wrap(err, "failed to get ads")
}
