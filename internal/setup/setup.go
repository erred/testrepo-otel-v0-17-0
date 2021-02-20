package setup

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InstallOtlpPipeline(ctx context.Context) (func(), error) {
	exporter, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("otel-collector.otel.svc.cluster.local:55680"),
	))
	if err != nil {
		return nil, fmt.Errorf("install otlp: create exporter: %w", err)
	}

	// 	semconv.ServiceNameKey.String("service-c"),
	// 	label.String("app", "svcc"),
	// ),
	// if err != nil {
	// 	return nil, fmt.Errorf("install otlp: create resource: %w", err)
	// }

	traceProvider := sdktrace.NewTracerProvider(sdktrace.WithConfig(
		sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample(),
		},
	), sdktrace.WithBatcher(
		exporter,
		// add following two options to ensure flush
		sdktrace.WithBatchTimeout(5),
		sdktrace.WithMaxExportBatchSize(10),
	))
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.Baggage{},
		propagation.TraceContext{},
	))

	return func() {
		ctx := context.TODO()
		err := traceProvider.Shutdown(ctx)
		if err != nil {
			otel.Handle(err)
		}
		err = exporter.Shutdown(ctx)
		if err != nil {
			otel.Handle(err)
		}
	}, nil
}
