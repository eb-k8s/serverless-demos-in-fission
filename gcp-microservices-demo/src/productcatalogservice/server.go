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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/productcatalogservice/rest"

	"github.com/sirupsen/logrus"
)

var (
	cat rest.ListProductsResponse
	log *logrus.Logger
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}
	log.Out = os.Stdout
	err := getCatalogData(&cat)
	if err != nil {
		log.Warnf("could not parse product catalog")
	}
}

// Handler is the entry point for this fission function
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Errorf("methods other than GET are not supported")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	if len(v) == 0 { //empty
		result, err := ListProducts()
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := json.Marshal(result)
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
	} else if v.Has("id") && !v.Has("query") {
		if len(v["id"]) > 1 {
			log.Errorf("could not get more than one product")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err := GetProduct(&rest.GetProductRequest{Id: v.Get("id")})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := json.Marshal(result)
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
	} else if !v.Has("id") && v.Has("query") {
		if len(v["query"]) > 1 {
			log.Errorf("could not search more than one product")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result, err := SearchProducts(&rest.SearchProductsRequest{Query: v.Get("query")})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := json.Marshal(result)
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
		log.Errorf("parameters of method GET are incorrect")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getCatalogData(catalog *rest.ListProductsResponse) error {
	if err := json.Unmarshal([]byte(productData), catalog); err != nil {
		log.Warnf("failed to parse the catalog JSON: %v", err)
		return err
	}
	log.Info("successfully parsed product catalog json")
	return nil
}

func parseCatalog() []*rest.Product {
	if len(cat.Products) == 0 {
		err := getCatalogData(&cat)
		if err != nil {
			return []*rest.Product{}
		}
	}
	return cat.Products
}

func ListProducts() (*rest.ListProductsResponse, error) {
	return &rest.ListProductsResponse{Products: parseCatalog()}, nil
}

func GetProduct(req *rest.GetProductRequest) (*rest.Product, error) {
	var found *rest.Product
	for i := 0; i < len(parseCatalog()); i++ {
		if req.Id == parseCatalog()[i].Id {
			found = parseCatalog()[i]
		}
	}
	if found == nil {
		return nil, fmt.Errorf("no product with ID %s", req.Id)
	}
	return found, nil
}

func SearchProducts(req *rest.SearchProductsRequest) (*rest.SearchProductsResponse, error) {
	// Intepret query as a substring match in name or description.
	var ps []*rest.Product
	for _, p := range parseCatalog() {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(p.Description), strings.ToLower(req.Query)) {
			ps = append(ps, p)
		}
	}
	return &rest.SearchProductsResponse{Results: ps}, nil
}
