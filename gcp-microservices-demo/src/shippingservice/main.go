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
	"os"
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

var log *logrus.Logger

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() {
	ctx := context.Background()
	// Get Resource
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("shippingservice"))

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
	log.Infof("adservice with opentelemetry collector: %s\n", otel_collector_addr)
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

	if r.Method == "POST" {
		span.AddEvent("invoke GetQuote")
		raw_req, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetQuote")
			return
		}
		req := new(GetQuoteRequest)
		err = json.Unmarshal(raw_req, req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetQuote")
			return
		}
		res, err := GetQuote(req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetQuote")
			return
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetQuote")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetQuote")
			return
		}
	} else if r.Method == "PUT" {
		span.AddEvent("invoke ShipOrder")
		raw_req, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ShipOrder")
			return
		}
		req := new(ShipOrderRequest)
		err = json.Unmarshal(raw_req, req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ShipOrder")
			return
		}
		res, err := ShipOrder(req)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ShipOrder")
			return
		}
		body, err := json.Marshal(res)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ShipOrder")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in ShipOrder")
			return
		}
	} else {
		log.Errorf("methods other than POST and PUT are not supported")
		w.WriteHeader(http.StatusBadRequest)
		span.AddEvent("methods other than POST and PUT are not supported")
		return
	}
	span.AddEvent("successfully handle request")
}

// GetQuote produces a shipping quote (cost) in USD.
func GetQuote(in *GetQuoteRequest) (*GetQuoteResponse, error) {
	log.Info("[GetQuote] received request")
	defer log.Info("[GetQuote] completed request")

	// 1. Generate a quote based on the total number of items to be shipped.
	quote := CreateQuoteFromCount(0)

	// 2. Generate a response.
	return &GetQuoteResponse{
		CostUsd: &Money{
			CurrencyCode: "USD",
			Units:        int64(quote.Dollars),
			Nanos:        int32(quote.Cents * 10000000)},
	}, nil

}

// ShipOrder mocks that the requested items will be shipped.
// It supplies a tracking ID for notional lookup of shipment delivery status.
func ShipOrder(in *ShipOrderRequest) (*ShipOrderResponse, error) {
	log.Info("[ShipOrder] received request")
	defer log.Info("[ShipOrder] completed request")
	// 1. Create a Tracking ID
	baseAddress := fmt.Sprintf("%s, %s, %s", in.Address.StreetAddress, in.Address.City, in.Address.State)
	id := CreateTrackingId(baseAddress)

	// 2. Generate a response.
	return &ShipOrderResponse{
		TrackingId: id,
	}, nil
}
