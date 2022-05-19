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
	"testing"
)

// TestGetQuote is a basic check on the GetQuote RPC service.
func TestGetQuote(t *testing.T) {
	// A basic test case to test logic and protobuf interactions.
	req := &GetQuoteRequest{
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

	res, err := GetQuote(req)
	if err != nil {
		t.Errorf("TestGetQuote (%v) failed", err)
	}
	if res.CostUsd.GetUnits() != 8 || res.CostUsd.GetNanos() != 990000000 {
		t.Errorf("TestGetQuote: Quote value '%d.%d' does not match expected '%s'", res.CostUsd.GetUnits(), res.CostUsd.GetNanos(), "11.220000000")
	}
}

// TestShipOrder is a basic check on the ShipOrder RPC service.
func TestShipOrder(t *testing.T) {
	// A basic test case to test logic and protobuf interactions.
	req := &ShipOrderRequest{
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

	res, err := ShipOrder(req)
	if err != nil {
		t.Errorf("TestShipOrder (%v) failed", err)
	}
	// @todo improve quality of this test to check for a pattern such as '[A-Z]{2}-\d+-\d+'.
	if len(res.TrackingId) != 18 {
		t.Errorf("TestShipOrder: Tracking ID is malformed - has %d characters, %d expected", len(res.TrackingId), 18)
	}
}
