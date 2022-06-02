package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/adservice/rest"
	"github.com/google/go-cmp/cmp"
)

func TestOneKey(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/ad?context_keys=clothing", nil)
	resp := httptest.NewRecorder()
	want := &rest.AdResponse{Ads: []*rest.Ad{
		{RedirectUrl: "/product/66VCHSJNUP", Text: "Tank top for sale. 20% off."},
	}}
	got := new(rest.AdResponse)
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

func TestEmptyKey(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/ad?context_keys=", nil)
	resp := httptest.NewRecorder()
	got := new(rest.AdResponse)
	Handler(resp, req)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(body, got)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.GetAds()) == 0 {
		t.Errorf("ads should be non-empty")
	}
}

func TestMoreKeys(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/ad?context_keys=footwear,decor", nil)
	resp := httptest.NewRecorder()
	want := &rest.AdResponse{Ads: []*rest.Ad{
		{RedirectUrl: "/product/L9ECAV7KIM", Text: "Loafers for sale. Buy one, get second one for free"},
		{RedirectUrl: "/product/0PUK6V6EV0", Text: "Candle holder for sale. 30% off."},
	}}
	got := new(rest.AdResponse)
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

func TestWrongKey(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/ad?wrong_key=decor", nil)
	resp := httptest.NewRecorder()
	Handler(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("should be bad request")
	}
}

func TestEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/ad", nil)
	resp := httptest.NewRecorder()
	Handler(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("should be bad request")
	}
}
