package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eb-k8s/serverless-demos-in-fission/gcp-microservices-demo/src/adservice/rest"
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

const MAX_ADS_TO_SERVE int = 2

var (
	adservice *Adservice
	log       *logrus.Logger
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() {
	ctx := context.Background()
	// Get Resource
	res := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("adservice"))

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
	rand.Seed(time.Now().UnixNano())

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

	adservice = new(Adservice)
	adservice.Set_max_ads_to_serve(MAX_ADS_TO_SERVE)
	adservice.CreateAdsMap()

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
	span.AddEvent("invoke GetAds")
	v := r.URL.Query()
	if v.Has("context_keys") {
		result, err := adservice.GetAds(&rest.AdRequest{ContextKeys: strings.Split(v.Get("context_keys"), ",")})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetAds")
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetAds")
			return
		}
		w.Header().Set("content-type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			span.AddEvent("an error occurred in GetAds")
			return
		}
	} else {
		log.Errorf("cannot get context_keys")
		w.WriteHeader(http.StatusBadRequest)
		span.AddEvent("cannot get context_keys")
		return
	}
	span.AddEvent("successfully handle request")
}

type Adservice struct {
	max_ads_to_serve int
	adsMap           map[string][]*rest.Ad
}

func (as *Adservice) GetAds(req *rest.AdRequest) (*rest.AdResponse, error) {
	var allads []*rest.Ad
	log.Info("received ad request (context_words=" + req.GetContextKeysList() + ")")
	if req.GetContextKeysCount() > 0 {
		for i := 0; i < req.GetContextKeysCount(); i++ {
			ads := as.GetAdsByCategory(req.GetContextkeys()[i])
			allads = append(allads, ads...)
		}
	} else {
		log.Info("No Context provided. Constructing random Ads.")
		allads = as.GetRandomAds()
	}
	if len(allads) == 0 {
		// Serve random ads.
		log.Info("No Ads found based on context. Constructing random Ads.")
		allads = as.GetRandomAds()
	}
	return &rest.AdResponse{Ads: allads}, nil
}

func (as *Adservice) CreateAdsMap() {
	as.adsMap = make(map[string][]*rest.Ad)
	hairdryer := &rest.Ad{RedirectUrl: "/product/2ZYFJ3GM2N", Text: "Hairdryer for sale. 50% off."}
	tankTop := &rest.Ad{RedirectUrl: "/product/66VCHSJNUP", Text: "Tank top for sale. 20% off."}
	candleHolder := &rest.Ad{RedirectUrl: "/product/0PUK6V6EV0", Text: "Candle holder for sale. 30% off."}
	bambooGlassJar := &rest.Ad{RedirectUrl: "/product/9SIQT8TOJO", Text: "Bamboo glass jar for sale. 10% off."}
	watch := &rest.Ad{RedirectUrl: "/product/1YMWWN1N4O", Text: "Watch for sale. Buy one, get second kit for free"}
	mug := &rest.Ad{RedirectUrl: "/product/6E92ZMYYFZ", Text: "Mug for sale. Buy two, get third one for free"}
	loafers := &rest.Ad{RedirectUrl: "/product/L9ECAV7KIM", Text: "Loafers for sale. Buy one, get second one for free"}
	as.adsMap["clothing"] = []*rest.Ad{tankTop}
	as.adsMap["accessories"] = []*rest.Ad{watch}
	as.adsMap["footwear"] = []*rest.Ad{loafers}
	as.adsMap["hair"] = []*rest.Ad{hairdryer}
	as.adsMap["decor"] = []*rest.Ad{candleHolder}
	as.adsMap["kitchen"] = []*rest.Ad{bambooGlassJar, mug}
}

func (as *Adservice) Set_max_ads_to_serve(max_ads_to_serve int) {
	as.max_ads_to_serve = max_ads_to_serve
}

func (as *Adservice) Get_max_ads_to_serve() int {
	return as.max_ads_to_serve
}

func (as *Adservice) GetAdsByCategory(category string) []*rest.Ad {
	return as.adsMap[category]
}

func (as *Adservice) GetRandomAds() []*rest.Ad {
	ads := make([]*rest.Ad, as.max_ads_to_serve)
	var allads []*rest.Ad
	for _, v := range as.adsMap {
		allads = append(allads, v...)
	}
	for i := 0; i < as.max_ads_to_serve; i++ {
		ads[i] = allads[rand.Intn(len(allads))]
	}
	return ads
}
