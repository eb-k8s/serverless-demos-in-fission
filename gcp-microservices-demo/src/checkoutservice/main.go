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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/checkoutservice/money"
	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/checkoutservice/rest"
)

const (
	productCatalogSvcAddr = "http://router.fission.svc.cluster.local/product"
	cartSvcAddr           = "http://router.fission.svc.cluster.local/cart"
	currencySvcAddr       = "http://router.fission.svc.cluster.local/currency"
	shippingSvcAddr       = "http://router.fission.svc.cluster.local/shipping"
	emailSvcAddr          = "http://router.fission.svc.cluster.local/email"
	paymentSvcAddr        = "http://router.fission.svc.cluster.local/payment"
)

var log *logrus.Logger
var svc *checkoutService

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
	svc = &checkoutService{
		productCatalogSvcAddr: productCatalogSvcAddr,
		cartSvcAddr:           cartSvcAddr,
		currencySvcAddr:       currencySvcAddr,
		shippingSvcAddr:       shippingSvcAddr,
		emailSvcAddr:          emailSvcAddr,
		paymentSvcAddr:        paymentSvcAddr,
	}
}

// Handler is the entry point for this fission function
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		raw_req, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req := new(rest.PlaceOrderRequest)
		err = json.Unmarshal(raw_req, req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		res, err := svc.PlaceOrder(req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		log.Errorf("methods other than POST are not supported")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

type checkoutService struct {
	productCatalogSvcAddr string
	cartSvcAddr           string
	currencySvcAddr       string
	shippingSvcAddr       string
	emailSvcAddr          string
	paymentSvcAddr        string
}

func (cs *checkoutService) PlaceOrder(req *rest.PlaceOrderRequest) (*rest.PlaceOrderResponse, error) {
	log.Infof("[PlaceOrder] user_id=%q user_currency=%q", req.UserId, req.UserCurrency)

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order uuid")
	}

	prep, err := cs.prepareOrderItemsAndShippingQuoteFromCart(req.UserId, req.UserCurrency, req.Address)
	if err != nil {
		return nil, err
	}

	total := rest.Money{CurrencyCode: req.UserCurrency, Units: 0, Nanos: 0}
	total = money.Must(money.Sum(total, *prep.shippingCostLocalized))
	for _, it := range prep.orderItems {
		multPrice := money.MultiplySlow(*it.Cost, uint32(it.GetItem().GetQuantity()))
		total = money.Must(money.Sum(total, multPrice))
	}

	txID, err := cs.chargeCard(&total, req.CreditCard)
	if err != nil {
		return nil, fmt.Errorf("failed to charge card: %+v", err)
	}
	log.Infof("payment went through (transaction_id: %s)", txID)

	shippingTrackingID, err := cs.shipOrder(req.Address, prep.cartItems)
	if err != nil {
		return nil, fmt.Errorf("shipping error: %+v", err)
	}

	err = cs.emptyUserCart(req.UserId)
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

	if err := cs.sendOrderConfirmation(req.Email, orderResult); err != nil {
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

func (cs *checkoutService) prepareOrderItemsAndShippingQuoteFromCart(userID, userCurrency string, address *rest.Address) (orderPrep, error) {
	var out orderPrep
	cartItems, err := cs.getUserCart(userID)
	if err != nil {
		return out, fmt.Errorf("cart failure: %+v", err)
	}
	orderItems, err := cs.prepOrderItems(cartItems, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to prepare order: %+v", err)
	}
	shippingUSD, err := cs.quoteShipping(address, cartItems)
	if err != nil {
		return out, fmt.Errorf("shipping quote failure: %+v", err)
	}
	shippingPrice, err := cs.convertCurrency(shippingUSD, userCurrency)
	if err != nil {
		return out, fmt.Errorf("failed to convert shipping cost to currency: %+v", err)
	}

	out.shippingCostLocalized = shippingPrice
	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (cs *checkoutService) quoteShipping(address *rest.Address, items []*rest.CartItem) (*rest.Money, error) {
	shippingQuote, err := rest.GetQuote(cs.shippingSvcAddr, &rest.GetQuoteRequest{
		Address: address,
		Items:   items,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping quote: %+v", err)
	}
	return shippingQuote.GetCostUsd(), nil
}

func (cs *checkoutService) getUserCart(userID string) ([]*rest.CartItem, error) {
	cart, err := rest.GetCart(cs.cartSvcAddr, &rest.GetCartRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart during checkout: %+v", err)
	}
	return cart.GetItems(), nil
}

func (cs *checkoutService) emptyUserCart(userID string) error {
	if err := rest.EmptyCart(cs.cartSvcAddr, &rest.EmptyCartRequest{UserId: userID}); err != nil {
		return fmt.Errorf("failed to empty user cart during checkout: %+v", err)
	}
	return nil
}

func (cs *checkoutService) prepOrderItems(items []*rest.CartItem, userCurrency string) ([]*rest.OrderItem, error) {
	out := make([]*rest.OrderItem, len(items))

	for i, item := range items {
		product, err := rest.GetProduct(cs.productCatalogSvcAddr, &rest.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			return nil, fmt.Errorf("failed to get product #%q", item.GetProductId())
		}
		price, err := cs.convertCurrency(product.GetPriceUsd(), userCurrency)
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

func (cs *checkoutService) convertCurrency(from *rest.Money, toCurrency string) (*rest.Money, error) {
	result, err := rest.Convert(cs.currencySvcAddr, &rest.CurrencyConversionRequest{
		From:   from,
		ToCode: toCurrency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %+v", err)
	}
	return result, err
}

func (cs *checkoutService) chargeCard(amount *rest.Money, paymentInfo *rest.CreditCardInfo) (string, error) {
	paymentResp, err := rest.Charge(cs.paymentSvcAddr, &rest.ChargeRequest{
		Amount:     amount,
		CreditCard: paymentInfo,
	})
	if err != nil {
		return "", fmt.Errorf("could not charge the card: %+v", err)
	}
	return paymentResp.GetTransactionId(), nil
}

func (cs *checkoutService) sendOrderConfirmation(email string, order *rest.OrderResult) error {
	err := rest.SendOrderConfirmation(cs.emailSvcAddr, &rest.SendOrderConfirmationRequest{
		Email: email,
		Order: order,
	})
	return err
}

func (cs *checkoutService) shipOrder(address *rest.Address, items []*rest.CartItem) (string, error) {
	resp, err := rest.ShipOrder(cs.shippingSvcAddr, &rest.ShipOrderRequest{
		Address: address,
		Items:   items,
	})
	if err != nil {
		return "", fmt.Errorf("shipment failed: %+v", err)
	}
	return resp.GetTrackingId(), nil
}
