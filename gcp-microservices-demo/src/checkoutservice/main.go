// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/checkoutservice/money"
	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/checkoutservice/rest"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type checkoutService struct {
	httpClient            http.Client
	productCatalogSvcAddr string
	cartSvcAddr           string
	currencySvcAddr       string
	shippingSvcAddr       string
	emailSvcAddr          string
	paymentSvcAddr        string
}

var log *logrus.Logger
var svc *checkoutService
var withOtel bool

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() {
	ctx := context.Background()
	// Get Resource
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("checkoutservice"))

	// Get Exporter
	traceExporter, err := getTraceExporter(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create trace exporter", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	if traceExporter != nil {
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tracerProvider.RegisterSpanProcessor(bsp)
	}
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	propagators := []propagation.TextMapPropagator{
		propagation.TraceContext{},
		propagation.Baggage{},
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagators...))
}

func getTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	otel_collector_addr := os.Getenv("OTEL_COLLECTOR_ADDR")
	if otel_collector_addr == "" {
		log.Info("OTEL_COLLECTOR_ADDR not set, skipping Opentelemtry tracing")
		withOtel = false
		return nil, nil
	}
	log.Infof("adservice with opentelemetry collector: %s\n", otel_collector_addr)
	withOtel = true
	grpcOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(otel_collector_addr),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
		otlptracegrpc.WithInsecure(),
	}
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, grpcOpts...)
	if err != nil {
		return nil, err
	}
	return traceExporter, nil
}

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Info("DOMAIN not set, skipping communicating with other functions")
		svc = &checkoutService{}
	} else {
		log.Infof("checkoutservice with domain: %s\n", domain)
		baseUrl, err := url.Parse(domain)
		if err != nil {
			log.Fatalf("%s: %s", "Malformed URL: ", err.Error())
		}
		baseUrl.Scheme = "http"
		productCatalogSvcUrl := *baseUrl
		productCatalogSvcUrl.Path += "/product"
		cartSvcUrl := *baseUrl
		cartSvcUrl.Path += "/cart"
		currencySvcUrl := *baseUrl
		currencySvcUrl.Path += "/currency"
		shippingSvcUrl := *baseUrl
		shippingSvcUrl.Path += "/shipping"
		emailSvcUrl := *baseUrl
		emailSvcUrl.Path += "/email"
		paymentSvcUrl := *baseUrl
		paymentSvcUrl.Path += "/payment"
		svc = &checkoutService{
			productCatalogSvcAddr: productCatalogSvcUrl.String(),
			cartSvcAddr:           cartSvcUrl.String(),
			currencySvcAddr:       currencySvcUrl.String(),
			shippingSvcAddr:       shippingSvcUrl.String(),
			emailSvcAddr:          emailSvcUrl.String(),
			paymentSvcAddr:        paymentSvcUrl.String(),
		}
	}

	initProvider()
	svc.httpClient = http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
}

// Handler is the entry point for this fission function
func Handler(w http.ResponseWriter, r *http.Request) {
	var tracer trace.Tracer
	if span := trace.SpanFromContext(r.Context()); span.SpanContext().IsValid() {
		tracer = span.TracerProvider().Tracer("")
	} else {
		tracer = otel.GetTracerProvider().Tracer("")
	}
	// Extract context from carrier
	propagators := otel.GetTextMapPropagator()
	ctx := propagators.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"handle request",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if r.Method == "POST" {
		span.AddEvent("invoke PlaceOrder")
		raw_req, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in PlaceOrder")
			return
		}
		req := new(rest.PlaceOrderRequest)
		err = json.Unmarshal(raw_req, req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in PlaceOrder")
			return
		}
		res, err := svc.PlaceOrder(ctx, tracer, req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in PlaceOrder")
			return
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in PlaceOrder")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in PlaceOrder")
			return
		}
	} else {
		log.Errorf("methods other than POST are not supported")
		w.WriteHeader(http.StatusBadRequest)
		span.AddEvent("methods other than POST are not supported")
		return
	}
	span.AddEvent("successfully handle request")
}

func (cs *checkoutService) PlaceOrder(ctx context.Context, tracer trace.Tracer, req *rest.PlaceOrderRequest) (*rest.PlaceOrderResponse, error) {
	log.Infof("[PlaceOrder] user_id=%q user_currency=%q", req.UserId, req.UserCurrency)

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order uuid")
	}

	prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(ctx, tracer, req.UserId, req.UserCurrency, req.Address)
	if err != nil {
		return nil, err
	}

	total := rest.Money{CurrencyCode: req.UserCurrency, Units: 0, Nanos: 0}
	total = money.Must(money.Sum(total, *prep.shippingCostLocalized))
	for _, it := range prep.orderItems {
		multPrice := money.MultiplySlow(*it.Cost, uint32(it.GetItem().GetQuantity()))
		total = money.Must(money.Sum(total, multPrice))
	}

	txID, err := cs.chargeCard(ctx, tracer, &total, req.CreditCard)
	if err != nil {
		return nil, fmt.Errorf("failed to charge card: %+v", err)
	}
	log.Infof("payment went through (transaction_id: %s)", txID)

	shippingTrackingID, err := cs.shipOrder(ctx, tracer, req.Address, prep.cartItems)
	if err != nil {
		return nil, fmt.Errorf("shipping error: %+v", err)
	}

	err = cs.emptyUserCart(ctx, tracer, req.UserId)
	if err != nil {
		return nil, err
	}

	orderResult := &rest.OrderResult{
		OrderId:            orderID.String(),
		ShippingTrackingId: shippingTrackingID,
		ShippingCost:       prep.shippingCostLocalized,
		ShippingAddress:    req.Address,
		Items:              prep.orderItems,
	}

	if err := cs.sendOrderConfirmation(ctx, tracer, req.Email, orderResult); err != nil {
		log.Warnf("failed to send order confirmation to %q: %+v", req.Email, err)
	} else {
		log.Infof("order confirmation email sent to %q", req.Email)
	}
	resp := &rest.PlaceOrderResponse{Order: orderResult}
	return resp, nil
}

type orderPrep struct {
	orderItems            []*rest.OrderItem
	cartItems             []*rest.CartItem
	shippingCostLocalized *rest.Money
}

func (cs *checkoutService) prepareOrderItemsAndShippingQuoteFromCart(ctx context.Context, tracer trace.Tracer, userID, userCurrency string, address *rest.Address) (orderPrep, error) {
	var out orderPrep

	cartItems, err := cs.getUserCart(ctx, tracer, userID)
	if err != nil {
		return out, fmt.Errorf("cart failure: %+v", err)
	}
	orderItems, err := cs.prepOrderItems(ctx, tracer, cartItems, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to prepare order: %+v", err)
	}
	shippingUSD, err := cs.quoteShipping(ctx, tracer, address, cartItems)
	if err != nil {
		return out, fmt.Errorf("shipping quote failure: %+v", err)
	}
	shippingPrice, err := cs.convertCurrency(ctx, tracer, shippingUSD, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to convert shipping cost to currency: %+v", err)
	}

	out.shippingCostLocalized = shippingPrice
	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (cs *checkoutService) quoteShipping(ctx context.Context, tracer trace.Tracer, address *rest.Address, items []*rest.CartItem) (*rest.Money, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetQuote",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetQuote")
	shippingQuote, err := rest.GetQuote(ctx, cs.httpClient, withOtel, cs.shippingSvcAddr, &rest.GetQuoteRequest{
		Address: address,
		Items:   items,
	})
	if err != nil {
		span.AddEvent("an error occurred in GetQuote")
		return nil, fmt.Errorf("failed to get shipping quote: %+v", err)
	}
	span.AddEvent("successfully invoke GetQuote")
	return shippingQuote.GetCostUsd(), nil
}

func (cs *checkoutService) getUserCart(ctx context.Context, tracer trace.Tracer, userID string) ([]*rest.CartItem, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke GetCart",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke GetCart")
	cart, err := rest.GetCart(ctx, cs.httpClient, withOtel, cs.cartSvcAddr, &rest.GetCartRequest{UserId: userID})
	if err != nil {
		span.AddEvent("an error occurred in GetCart")
		return nil, fmt.Errorf("failed to get user cart during checkout: %+v", err)
	}
	span.AddEvent("successfully invoke GetCart")
	return cart.GetItems(), nil
}

func (cs *checkoutService) emptyUserCart(ctx context.Context, tracer trace.Tracer, userID string) error {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke EmptyCart",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke EmptyCart")
	if err := rest.EmptyCart(ctx, cs.httpClient, withOtel, cs.cartSvcAddr, &rest.EmptyCartRequest{UserId: userID}); err != nil {
		span.AddEvent("an error occurred in EmptyCart")
		return fmt.Errorf("failed to empty user cart during checkout: %+v", err)
	}
	span.AddEvent("successfully invoke EmptyCart")
	return nil
}

func (cs *checkoutService) prepOrderItems(ctx context.Context, tracer trace.Tracer, items []*rest.CartItem, userCurrency string) ([]*rest.OrderItem, error) {
	out := make([]*rest.OrderItem, len(items))

	for i, item := range items {
		// Start a span
		ctx_getproduct, span := tracer.Start(
			ctx,
			"invoke GetProduct",
			trace.WithSpanKind(trace.SpanKindClient),
		)

		span.AddEvent("invoke GetProduct")
		product, err := rest.GetProduct(ctx_getproduct, cs.httpClient, withOtel, cs.productCatalogSvcAddr, &rest.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			span.AddEvent("an error occurred in GetProduct")
			span.End()
			return nil, fmt.Errorf("failed to get product #%q", item.GetProductId())
		}
		span.AddEvent("successfully invoke GetProduct")
		span.End()

		price, err := cs.convertCurrency(ctx, tracer, product.GetPriceUsd(), userCurrency)
		if err != nil {
			return nil, fmt.Errorf("failed to convert price of %q to %s", item.GetProductId(), userCurrency)
		}
		out[i] = &rest.OrderItem{
			Item: item,
			Cost: price,
		}
	}
	return out, nil
}

func (cs *checkoutService) convertCurrency(ctx context.Context, tracer trace.Tracer, from *rest.Money, toCurrency string) (*rest.Money, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke Convert",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke Convert")
	result, err := rest.Convert(ctx, cs.httpClient, withOtel, cs.currencySvcAddr, &rest.CurrencyConversionRequest{
		From:   from,
		ToCode: toCurrency,
	})
	if err != nil {
		span.AddEvent("an error occurred in Convert")
		return nil, fmt.Errorf("failed to convert currency: %+v", err)
	}
	span.AddEvent("successfully invoke Convert")
	return result, err
}

func (cs *checkoutService) chargeCard(ctx context.Context, tracer trace.Tracer, amount *rest.Money, paymentInfo *rest.CreditCardInfo) (string, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke Charge",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke Charge")
	paymentResp, err := rest.Charge(ctx, cs.httpClient, withOtel, cs.paymentSvcAddr, &rest.ChargeRequest{
		Amount:     amount,
		CreditCard: paymentInfo,
	})
	if err != nil {
		span.AddEvent("an error occurred in Charge")
		return "", fmt.Errorf("could not charge the card: %+v", err)
	}
	span.AddEvent("successfully invoke Charge")
	return paymentResp.GetTransactionId(), nil
}

func (cs *checkoutService) sendOrderConfirmation(ctx context.Context, tracer trace.Tracer, email string, order *rest.OrderResult) error {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke SendOrderConfirmation",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke SendOrderConfirmation")
	err := rest.SendOrderConfirmation(ctx, cs.httpClient, withOtel, cs.emailSvcAddr, &rest.SendOrderConfirmationRequest{
		Email: email,
		Order: order,
	})
	if err != nil {
		span.AddEvent("an error occurred in SendOrderConfirmation")
		return err
	}
	span.AddEvent("successfully invoke SendOrderConfirmation")
	return nil
}

func (cs *checkoutService) shipOrder(ctx context.Context, tracer trace.Tracer, address *rest.Address, items []*rest.CartItem) (string, error) {
	// Start a span
	ctx, span := tracer.Start(
		ctx,
		"invoke ShipOrder",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	span.AddEvent("invoke ShipOrder")
	resp, err := rest.ShipOrder(ctx, cs.httpClient, withOtel, cs.shippingSvcAddr, &rest.ShipOrderRequest{
		Address: address,
		Items:   items,
	})
	if err != nil {
		span.AddEvent("an error occurred in ShipOrder")
		return "", fmt.Errorf("shipment failed: %+v", err)
	}
	span.AddEvent("successfully invoke ShipOrder")
	return resp.GetTrackingId(), nil
}
