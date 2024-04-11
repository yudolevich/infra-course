package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ServiceName = "service"
)

var (
	prv *sdktrace.TracerProvider
	tr trace.Tracer
	errSpan = errors.New("span error")
)

func initTracer(ctx context.Context) error{
	conn, err := grpc.NewClient("jaeger:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
		),
	)
	if err != nil {
		return err
	}

	prv = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(prv)
	tr = prv.Tracer("tracer")

	return nil
}

func main() {
	ctx := context.Background()
	if err := initTracer(ctx);err != nil {
		log.Fatal("init tracer", err)
	}

	ctx, span := tr.Start(ctx, "root span")

	test(ctx, 1)
	
	test(ctx, 3)

	test(ctx, 1)

	span.End()

	if err := prv.Shutdown(ctx); err != nil {
		log.Fatal("failed shutdown", err)
	}
}

func test(ctx context.Context, count int) {
	if count < 1 {
		return
	}
	ctx, span := tr.Start(ctx, fmt.Sprintf("span-%d", count))
	defer span.End()
	test(ctx, count - 1)
	num := rand.Intn(5)+1
	time.Sleep(time.Duration(num)*time.Second)
	if num%2 == 0 {
		span.SetStatus(codes.Error, errSpan.Error())
		span.RecordError(errSpan)
	}
	log.Println("called test", count, num)
}
