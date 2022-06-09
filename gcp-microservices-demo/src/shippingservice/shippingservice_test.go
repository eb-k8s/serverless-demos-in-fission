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
	"bytes"
	"encoding/json"
	"io"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMethodNotPOSTOrPUT(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/shipping", nil)
	resp := httptest.NewRecorder()
	Handler(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("should be bad request")
	}
}

func TestWrongBody(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com/shipping", nil)
	resp := httptest.NewRecorder()
	Handler(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("should be bad request")
	}
}

func TestGetQuote(t *testing.T) {
	req_body := &GetQuoteRequest{
		Address: &Address{
			StreetAddress: "Muffin Man",
			City:          "London",
			State:         "",
			Country:       "England",
		},
		Items: []*CartItem{
			{
				ProductId: "23",
				Quantity:  1,
			},
			{
				ProductId: "46",
				Quantity:  3,
			},
		},
	}
	payload, err := json.Marshal(req_body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "http://example.com/shipping", bytes.NewBuffer(payload))
	resp := httptest.NewRecorder()
	want := &GetQuoteResponse{CostUsd: &Money{CurrencyCode: "USD", Units: 8, Nanos: 990000000}}
	got := new(GetQuoteResponse)
	Handler(resp, req)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(body, got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestShipOrder(t *testing.T) {
	req_body := &ShipOrderRequest{
		Address: &Address{
			StreetAddress: "Muffin Man",
			City:          "London",
			State:         "",
			Country:       "England",
		},
		Items: []*CartItem{
			{
				ProductId: "23",
				Quantity:  1,
			},
			{
				ProductId: "46",
				Quantity:  3,
			},
		},
	}
	payload, err := json.Marshal(req_body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("PUT", "http://example.com/shipping", bytes.NewBuffer(payload))
	resp := httptest.NewRecorder()
	got := new(ShipOrderResponse)
	Handler(resp, req)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(body, got)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TrackingId) != 18 {
		t.Errorf("TestShipOrder: Tracking ID is malformed - has %d characters, %d expected", len(got.TrackingId), 18)
	}
}
