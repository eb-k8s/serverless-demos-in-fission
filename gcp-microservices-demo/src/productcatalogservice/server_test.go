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

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/productcatalogservice/rest"

	"github.com/google/go-cmp/cmp"
)

func TestServer(t *testing.T) {
	res, err := ListProducts()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res.Products, parseCatalog()); diff != "" {
		t.Error(diff)
	}

	got, err := GetProduct(&rest.GetProductRequest{Id: "OLJCESPC7Z"})
	if err != nil {
		t.Fatal(err)
	}
	want := parseCatalog()[0]
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("got %v, want %v", got, want)
	}

	sres, err := SearchProducts(&rest.SearchProductsRequest{Query: "sunglasses"})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(sres.Results, []*rest.Product{parseCatalog()[0]}); diff != "" {
		t.Error(diff)
	}
}
