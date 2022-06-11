package main

import (
	"context"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/frontend/rest"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
)

func (fe *frontendServer) getCurrencies(ctx context.Context, tracer trace.Tracer) ([]string, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetSupportedCurrencies",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetSupportedCurrencies")
	currs, err := rest.GetSupportedCurrencies(ctx, fe.httpClient, fe.currencySvcAddr)
	if err != nil {
		span.AddEvent("an error occurred in GetSupportedCurrencies")
		return nil, err
	}
	var out []string
	for _, c := range currs.CurrencyCodes {
		if _, ok := whitelistedCurrencies[c]; ok {
			out = append(out, c)
		}
	}
	span.AddEvent("successfully invoke GetSupportedCurrencies")
	return out, nil
}

func (fe *frontendServer) getProducts(ctx context.Context, tracer trace.Tracer) ([]*rest.Product, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke ListProducts",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke ListProducts")
	resp, err := rest.ListProducts(ctx, fe.httpClient, fe.productCatalogSvcAddr)
	if err != nil {
		span.AddEvent("an error occurred in ListProducts")
		return resp.GetProducts(), err
	}
	span.AddEvent("successfully invoke ListProducts")
	return resp.GetProducts(), nil
}

func (fe *frontendServer) getProduct(ctx context.Context, tracer trace.Tracer, id string) (*rest.Product, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetProduct",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetProduct")
	resp, err := rest.GetProduct(ctx, fe.httpClient, fe.productCatalogSvcAddr, &rest.GetProductRequest{Id: id})
	if err != nil {
		span.AddEvent("an error occurred in GetProduct")
		return resp, err
	}
	span.AddEvent("successfully invoke GetProduct")
	return resp, nil
}

func (fe *frontendServer) getCart(ctx context.Context, tracer trace.Tracer, userID string) ([]*rest.CartItem, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetCart",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetCart")
	resp, err := rest.GetCart(ctx, fe.httpClient, fe.cartSvcAddr, &rest.GetCartRequest{UserId: userID})
	if err != nil {
		span.AddEvent("an error occurred in GetCart")
		return resp.GetItems(), err
	}
	span.AddEvent("successfully invoke GetCart")
	return resp.GetItems(), nil
}

func (fe *frontendServer) emptyCart(ctx context.Context, tracer trace.Tracer, userID string) error {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke EmptyCart",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke EmptyCart")
	err := rest.EmptyCart(ctx, fe.httpClient, fe.cartSvcAddr, &rest.EmptyCartRequest{UserId: userID})
	if err != nil {
		span.AddEvent("an error occurred in EmptyCart")
		return err
	}
	span.AddEvent("successfully invoke EmptyCart")
	return nil
}

func (fe *frontendServer) insertCart(ctx context.Context, tracer trace.Tracer, userID, productID string, quantity int32) error {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke AddItem",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke AddItem")
	err := rest.AddItem(ctx, fe.httpClient, fe.cartSvcAddr, &rest.AddItemRequest{
		UserId: userID,
		Item: &rest.CartItem{
			ProductId: productID,
			Quantity:  quantity},
	})
	if err != nil {
		span.AddEvent("an error occurred in AddItem")
		return err
	}
	span.AddEvent("successfully invoke AddItem")
	return nil
}

func (fe *frontendServer) convertCurrency(ctx context.Context, tracer trace.Tracer, money *rest.Money, currency string) (*rest.Money, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke Convert",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke Convert")
	resp, err := rest.Convert(ctx, fe.httpClient, fe.currencySvcAddr, &rest.CurrencyConversionRequest{
		From:   money,
		ToCode: currency})
	if err != nil {
		span.AddEvent("an error occurred in Convert")
		return resp, err
	}
	span.AddEvent("successfully invoke Convert")
	return resp, nil
}

func (fe *frontendServer) getShippingQuote(ctx context.Context, tracer trace.Tracer, items []*rest.CartItem, currency string) (*rest.Money, error) {
	// Start a span
	ctx_getquote, span := tracer.Start(
		ctx,
		"invoke GetQuote",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	span.AddEvent("invoke GetQuote")
	quote, err := rest.GetQuote(ctx_getquote, fe.httpClient, fe.shippingSvcAddr, &rest.GetQuoteRequest{
		Address: nil,
		Items:   items})
	if err != nil {
		span.AddEvent("an error occurred in GetQuote")
		span.End()
		return nil, err
	}
	span.AddEvent("successfully invoke GetQuote")
	span.End()

	localized, err := fe.convertCurrency(ctx, tracer, quote.GetCostUsd(), currency)
	return localized, errors.Wrap(err, "failed to convert currency for shipping cost")
}

func (fe *frontendServer) getRecommendations(ctx context.Context, tracer trace.Tracer, userID string, productIDs []string) ([]*rest.Product, error) {
	// Start a span
	ctx_listrecommendations, span := tracer.Start(
		ctx,
		"invoke ListRecommendations",
		trace.WithSpanKind(trace.SpanKindClient),
	)

	span.AddEvent("invoke ListRecommendations")
	resp, err := rest.ListRecommendations(ctx_listrecommendations, fe.httpClient, fe.recommendationSvcAddr, &rest.ListRecommendationsRequest{UserId: userID, ProductIds: productIDs})
	if err != nil {
		span.AddEvent("an error occurred in ListRecommendations")
		span.End()
		return nil, err
	}
	span.AddEvent("successfully invoke ListRecommendations")
	span.End()

	out := make([]*rest.Product, len(resp.GetProductIds()))
	for i, v := range resp.GetProductIds() {
		p, err := fe.getProduct(ctx, tracer, v)
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

func (fe *frontendServer) getAd(ctx context.Context, tracer trace.Tracer, ctxKeys []string) ([]*rest.Ad, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetAds",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetAds")
	resp, err := rest.GetAds(ctx, fe.httpClient, fe.adSvcAddr, &rest.AdRequest{
		ContextKeys: ctxKeys,
	})
	if err != nil {
		span.AddEvent("an error occurred in GetAds")
		return resp.GetAds(), errors.Wrap(err, "failed to get ads")
	}
	span.AddEvent("successfully invoke GetAds")
	return resp.GetAds(), nil
}
