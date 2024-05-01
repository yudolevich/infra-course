package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	service := os.Getenv("NAME")
	ctx := context.Background()

	conn, err := grpc.NewClient("collector:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("connect to collector", err)
	}

	tr, err := initTracer(ctx, conn, service)
	if err != nil {
		log.Fatal("init tracer", err)
	}

	_, err = initMeter(ctx, conn, service)
	if err != nil {
		log.Fatal("init meter", err)
	}

	http.Handle("/", newHandler(service, tr))
	http.ListenAndServe(":8080", nil)
}

func initTracer(ctx context.Context, conn *grpc.ClientConn, svc string) (trace.Tracer, error) {
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(svc),
		),
	)
	if err != nil {
		return nil, err
	}

	prv := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(prv)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return prv.Tracer("tracer"), nil
}

func initMeter(ctx context.Context, conn *grpc.ClientConn, svc string) (metric.MeterProvider, error) {
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(svc),
		),
	)
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
	)
	otel.SetMeterProvider(mp)

	return mp, nil
}

func sendReq(ctx context.Context, tr trace.Tracer, url string) error {
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response code: %d", resp.StatusCode)
	}

	return nil
}

func newHandler(name string, tr trace.Tracer) http.Handler {
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Printf("request %s\n", r.URL.Path)

		path := strings.Split(r.URL.Path, "/")
		if len(path) > 1 && len(path[1]) > 0 {
			log.Printf("send request to %s\n", path[1])
			if err := sendReq(ctx, tr,
				fmt.Sprintf(
					"http://%s:8080/%s", path[1], strings.Join(path[2:], "/"),
				)); err != nil {
				log.Printf("send request error %s", err)
				span := trace.SpanFromContext(ctx)
				span.SetStatus(codes.Error, "error span")
				span.RecordError(fmt.Errorf("error span"))
			}
		}

		num := rand.Intn(5) + 1
		time.Sleep(time.Duration(num) * time.Second)
		if num%3 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}), name)
}
