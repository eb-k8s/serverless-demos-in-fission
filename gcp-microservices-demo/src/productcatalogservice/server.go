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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	cat ListProductsResponse
	log *logrus.Logger
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() {
	ctx := context.Background()
	// Get Resource
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("productcatalogservice"))

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
		return nil, nil
	}
	log.Infof("productcatalogservice with opentelemetry collector: %s\n", otel_collector_addr)
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

	initProvider()
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
	_, span := tracer.Start(
		ctx,
		"handle request",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	if r.Method != "GET" {
		log.Errorf("methods other than GET are not supported")
		w.WriteHeader(http.StatusBadRequest)
		span.AddEvent("methods other than GET are not supported")
		return
	}
	v := r.URL.Query()
	if len(v) == 0 { //empty
		span.AddEvent("invoke ListProducts")
		result, err := ListProducts()
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ListProducts")
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ListProducts")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ListProducts")
			return
		}
	} else if v.Has("id") && !v.Has("query") {
		if len(v["id"]) > 1 {
			log.Errorf("could not get more than one product")
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("could not get more than one product")
			return
		}
		span.AddEvent("invoke GetProduct")
		result, err := GetProduct(&GetProductRequest{Id: v.Get("id")})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetProduct")
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetProduct")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetProduct")
			return
		}
	} else if !v.Has("id") && v.Has("query") {
		if len(v["query"]) > 1 {
			log.Errorf("could not search more than one product")
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("could not search more than one product")
			return
		}
		span.AddEvent("invoke SearchProducts")
		result, err := SearchProducts(&SearchProductsRequest{Query: v.Get("query")})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in SearchProducts")
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in SearchProducts")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in SearchProducts")
			return
		}
	} else {
		log.Errorf("parameters of method GET are incorrect")
		w.WriteHeader(http.StatusBadRequest)
		span.AddEvent("parameters of method GET are incorrect")
		return
	}
	span.AddEvent("successfully handle request")
}

func getCatalogData(catalog *ListProductsResponse) error {
	if err := json.Unmarshal([]byte(productData), catalog); err != nil {
		log.Warnf("failed to parse the catalog JSON: %v", err)
		return err
	}
	log.Info("successfully parsed product catalog json")
	return nil
}

func parseCatalog() []*Product {
	if len(cat.Products) == 0 {
		err := getCatalogData(&cat)
		if err != nil {
			return []*Product{}
		}
	}
	return cat.Products
}

func ListProducts() (*ListProductsResponse, error) {
	return &ListProductsResponse{Products: parseCatalog()}, nil
}

func GetProduct(req *GetProductRequest) (*Product, error) {
	var found *Product
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

func SearchProducts(req *SearchProductsRequest) (*SearchProductsResponse, error) {
	// Intepret query as a substring match in name or description.
	var ps []*Product
	for _, p := range parseCatalog() {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(p.Description), strings.ToLower(req.Query)) {
			ps = append(ps, p)
		}
	}
	return &SearchProductsResponse{Results: ps}, nil
}
