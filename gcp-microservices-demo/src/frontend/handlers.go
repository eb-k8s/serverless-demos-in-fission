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
	"fmt"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/frontend/money"
	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/frontend/rest"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type platformDetails struct {
	css      string
	provider string
}

var (
	isCymbalBrand = "true" == strings.ToLower(os.Getenv("CYMBAL_BRANDING"))
	templates     = template.Must(template.New("").
			Funcs(template.FuncMap{
			"renderMoney":        renderMoney,
			"renderCurrencyLogo": renderCurrencyLogo,
		}).ParseGlob("templates/*.html"))
	plat platformDetails
)

var validEnvs = []string{"local", "gcp", "azure", "aws", "onprem", "alibaba"}

func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
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
		"/",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.WithField("currency", currentCurrency(r)).Info("home")
	currencies, err := fe.getCurrencies(ctx, tracer)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	products, err := fe.getProducts(ctx, tracer)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve products"), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(ctx, tracer, sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	type productView struct {
		Item  *rest.Product
		Price *rest.Money
	}
	ps := make([]productView, len(products))
	for i, p := range products {
		price, err := fe.convertCurrency(ctx, tracer, p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "failed to do currency conversion for product %s", p.GetId()), http.StatusInternalServerError)
			return
		}
		ps[i] = productView{p, price}
	}

	// Set ENV_PLATFORM (default to local if not set; use env var if set; otherwise detect GCP, which overrides env)_
	var env = os.Getenv("ENV_PLATFORM")
	// Only override from env variable if set + valid env
	if env == "" || stringinSlice(validEnvs, env) == false {
		fmt.Println("env platform is either empty or invalid")
		env = "local"
	}
	// Autodetect GCP
	addrs, err := net.LookupHost("metadata.google.internal.")
	if err == nil && len(addrs) >= 0 {
		log.Debugf("Detected Google metadata server: %v, setting ENV_PLATFORM to GCP.", addrs)
		env = "gcp"
	}

	log.Debugf("ENV_PLATFORM is: %s", env)
	plat = platformDetails{}
	plat.setPlatformDetails(strings.ToLower(env))

	if err := templates.ExecuteTemplate(w, "home", map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"show_currency":     true,
		"currencies":        currencies,
		"products":          ps,
		"cart_size":         cartSize(cart),
		"banner_color":      os.Getenv("BANNER_COLOR"), // illustrates canary deployments
		"ad":                fe.chooseAd(ctx, tracer, []string{}, log),
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"deploymentDetails": deploymentDetailsMap,
	}); err != nil {
		log.Error(err)
	}
	span.AddEvent("successfully handle request")
}

func (plat *platformDetails) setPlatformDetails(env string) {
	if env == "aws" {
		plat.provider = "AWS"
		plat.css = "aws-platform"
	} else if env == "onprem" {
		plat.provider = "On-Premises"
		plat.css = "onprem-platform"
	} else if env == "azure" {
		plat.provider = "Azure"
		plat.css = "azure-platform"
	} else if env == "gcp" {
		plat.provider = "Google Cloud"
		plat.css = "gcp-platform"
	} else if env == "alibaba" {
		plat.provider = "Alibaba Cloud"
		plat.css = "alibaba-platform"
	} else {
		plat.provider = "local"
		plat.css = "local"
	}
}

func (fe *frontendServer) productHandler(w http.ResponseWriter, r *http.Request) {
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
		"/product/{id}",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	id := mux.Vars(r)["id"]
	if id == "" {
		renderHTTPError(log, r, w, errors.New("product id not specified"), http.StatusBadRequest)
		return
	}
	log.WithField("id", id).WithField("currency", currentCurrency(r)).
		Debug("serving product page")

	p, err := fe.getProduct(ctx, tracer, id)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}
	currencies, err := fe.getCurrencies(ctx, tracer)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}

	cart, err := fe.getCart(ctx, tracer, sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	price, err := fe.convertCurrency(ctx, tracer, p.GetPriceUsd(), currentCurrency(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to convert currency"), http.StatusInternalServerError)
		return
	}

	recommendations, err := fe.getRecommendations(ctx, tracer, sessionID(r), []string{id})
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get product recommendations"), http.StatusInternalServerError)
		return
	}

	product := struct {
		Item  *rest.Product
		Price *rest.Money
	}{p, price}

	if err := templates.ExecuteTemplate(w, "product", map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"ad":                fe.chooseAd(ctx, tracer, p.Categories, log),
		"user_currency":     currentCurrency(r),
		"show_currency":     true,
		"currencies":        currencies,
		"product":           product,
		"recommendations":   recommendations,
		"cart_size":         cartSize(cart),
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"deploymentDetails": deploymentDetailsMap,
	}); err != nil {
		log.Println(err)
	}
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
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
		"POST /cart",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
	productID := r.FormValue("product_id")
	if productID == "" || quantity == 0 {
		renderHTTPError(log, r, w, errors.New("invalid form input"), http.StatusBadRequest)
		return
	}
	log.WithField("product", productID).WithField("quantity", quantity).Debug("adding to cart")

	p, err := fe.getProduct(ctx, tracer, productID)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}

	if err := fe.insertCart(ctx, tracer, sessionID(r), p.GetId(), int32(quantity)); err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to add to cart"), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", "/cart")
	w.WriteHeader(http.StatusFound)
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) emptyCartHandler(w http.ResponseWriter, r *http.Request) {
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
		"/cart/empty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("emptying cart")

	if err := fe.emptyCart(ctx, tracer, sessionID(r)); err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to empty cart"), http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", "/")
	w.WriteHeader(http.StatusFound)
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) viewCartHandler(w http.ResponseWriter, r *http.Request) {
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
		"GET /cart",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("view user cart")
	currencies, err := fe.getCurrencies(ctx, tracer)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	cart, err := fe.getCart(ctx, tracer, sessionID(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}

	recommendations, err := fe.getRecommendations(ctx, tracer, sessionID(r), cartIDs(cart))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get product recommendations"), http.StatusInternalServerError)
		return
	}

	shippingCost, err := fe.getShippingQuote(ctx, tracer, cart, currentCurrency(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to get shipping quote"), http.StatusInternalServerError)
		return
	}

	type cartItemView struct {
		Item     *rest.Product
		Quantity int32
		Price    *rest.Money
	}
	items := make([]cartItemView, len(cart))
	totalPrice := rest.Money{CurrencyCode: currentCurrency(r)}
	for i, item := range cart {
		p, err := fe.getProduct(ctx, tracer, item.GetProductId())
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "could not retrieve product #%s", item.GetProductId()), http.StatusInternalServerError)
			return
		}
		price, err := fe.convertCurrency(ctx, tracer, p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "could not convert currency for product #%s", item.GetProductId()), http.StatusInternalServerError)
			return
		}

		multPrice := money.MultiplySlow(*price, uint32(item.GetQuantity()))
		items[i] = cartItemView{
			Item:     p,
			Quantity: item.GetQuantity(),
			Price:    &multPrice}
		totalPrice = money.Must(money.Sum(totalPrice, multPrice))
	}
	totalPrice = money.Must(money.Sum(totalPrice, *shippingCost))
	year := time.Now().Year()

	if err := templates.ExecuteTemplate(w, "cart", map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"currencies":        currencies,
		"recommendations":   recommendations,
		"cart_size":         cartSize(cart),
		"shipping_cost":     shippingCost,
		"show_currency":     true,
		"total_cost":        totalPrice,
		"items":             items,
		"expiration_years":  []int{year, year + 1, year + 2, year + 3, year + 4},
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"deploymentDetails": deploymentDetailsMap,
	}); err != nil {
		log.Println(err)
	}
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
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
		"/cart/checkout",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("placing order")

	var (
		email         = r.FormValue("email")
		streetAddress = r.FormValue("street_address")
		zipCode, _    = strconv.ParseInt(r.FormValue("zip_code"), 10, 32)
		city          = r.FormValue("city")
		state         = r.FormValue("state")
		country       = r.FormValue("country")
		ccNumber      = r.FormValue("credit_card_number")
		ccMonth, _    = strconv.ParseInt(r.FormValue("credit_card_expiration_month"), 10, 32)
		ccYear, _     = strconv.ParseInt(r.FormValue("credit_card_expiration_year"), 10, 32)
		ccCVV, _      = strconv.ParseInt(r.FormValue("credit_card_cvv"), 10, 32)
	)

	// Start a span
	ctx_placeorder, span_placeorder := tracer.Start(
		ctx,
		"invoke PlaceOrder",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	order, err := rest.PlaceOrder(ctx_placeorder, svc.httpClient, fe.checkoutSvcAddr, &rest.PlaceOrderRequest{
		Email: email,
		CreditCard: &rest.CreditCardInfo{
			CreditCardNumber:          ccNumber,
			CreditCardExpirationMonth: int32(ccMonth),
			CreditCardExpirationYear:  int32(ccYear),
			CreditCardCvv:             int32(ccCVV)},
		UserId:       sessionID(r),
		UserCurrency: currentCurrency(r),
		Address: &rest.Address{
			StreetAddress: streetAddress,
			City:          city,
			State:         state,
			ZipCode:       int32(zipCode),
			Country:       country},
	})
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to complete the order"), http.StatusInternalServerError)
		span_placeorder.AddEvent("an error occurred in PlaceOrder")
		span_placeorder.End()
		return
	}
	log.WithField("order", order.GetOrder().GetOrderId()).Info("order placed")
	span_placeorder.AddEvent("successfully invoke PlaceOrder")
	span_placeorder.End()

	order.GetOrder().GetItems()
	recommendations, _ := fe.getRecommendations(ctx, tracer, sessionID(r), nil)

	totalPaid := *order.GetOrder().GetShippingCost()
	for _, v := range order.GetOrder().GetItems() {
		multPrice := money.MultiplySlow(*v.GetCost(), uint32(v.GetItem().GetQuantity()))
		totalPaid = money.Must(money.Sum(totalPaid, multPrice))
	}

	currencies, err := fe.getCurrencies(ctx, tracer)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "order", map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"show_currency":     false,
		"currencies":        currencies,
		"order":             order.GetOrder(),
		"total_paid":        &totalPaid,
		"recommendations":   recommendations,
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"deploymentDetails": deploymentDetailsMap,
	}); err != nil {
		log.Println(err)
	}
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
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
		"/logout",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.Debug("logging out")
	for _, c := range r.Cookies() {
		c.Expires = time.Now().Add(-time.Hour * 24 * 365)
		c.MaxAge = -1
		http.SetCookie(w, c)
	}
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
	span.AddEvent("successfully handle request")
}

func (fe *frontendServer) setCurrencyHandler(w http.ResponseWriter, r *http.Request) {
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
		"/setCurrency",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	span.AddEvent("handle request")
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	cur := r.FormValue("currency_code")
	log.WithField("curr.new", cur).WithField("curr.old", currentCurrency(r)).
		Debug("setting currency")

	if cur != "" {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieCurrency,
			Value:  cur,
			MaxAge: cookieMaxAge,
		})
	}
	referer := r.Header.Get("referer")
	if referer == "" {
		referer = "/"
	}
	w.Header().Set("Location", referer)
	w.WriteHeader(http.StatusFound)
	span.AddEvent("successfully handle request")
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, tracer trace.Tracer, ctxKeys []string, log logrus.FieldLogger) *rest.Ad {
	ads, err := fe.getAd(ctx, tracer, ctxKeys)
	if err != nil {
		log.WithField("error", err).Warn("failed to retrieve ads")
		return nil
	}
	return ads[rand.Intn(len(ads))]
}

func renderHTTPError(log logrus.FieldLogger, r *http.Request, w http.ResponseWriter, err error, code int) {
	log.WithField("error", err).Error("request error")
	errMsg := fmt.Sprintf("%+v", err)

	w.WriteHeader(code)

	if templateErr := templates.ExecuteTemplate(w, "error", map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"error":             errMsg,
		"status_code":       code,
		"status":            http.StatusText(code),
		"deploymentDetails": deploymentDetailsMap,
	}); templateErr != nil {
		log.Println(templateErr)
	}
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

func cartIDs(c []*rest.CartItem) []string {
	out := make([]string, len(c))
	for i, v := range c {
		out[i] = v.GetProductId()
	}
	return out
}

// get total # of items in cart
func cartSize(c []*rest.CartItem) int {
	cartSize := 0
	for _, item := range c {
		cartSize += int(item.GetQuantity())
	}
	return cartSize
}

func renderMoney(money rest.Money) string {
	currencyLogo := renderCurrencyLogo(money.GetCurrencyCode())
	return fmt.Sprintf("%s%d.%02d", currencyLogo, money.GetUnits(), money.GetNanos()/10000000)
}

func renderCurrencyLogo(currencyCode string) string {
	logos := map[string]string{
		"USD": "$",
		"CAD": "$",
		"JPY": "¥",
		"EUR": "€",
		"TRY": "₺",
		"GBP": "£",
	}

	logo := "$" //default
	if val, ok := logos[currencyCode]; ok {
		logo = val
	}
	return logo
}

func stringinSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
