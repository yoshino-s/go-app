package otlp

import (
	"context"
	"time"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func (app *OtlpApp) configureMetrics(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	opts := app.config.metricOptions
	opts = append(opts, sdkmetric.WithResource(app.resource))
	dsn, err := ParseDSN(app.config.dsn)
	if err != nil {
		return nil, err
	}

	exp, err := otlpmetricClient(ctx, app.config, dsn)
	if err != nil {
		return nil, err
	}

	reader := sdkmetric.NewPeriodicReader(
		exp,
		sdkmetric.WithInterval(15*time.Second),
	)
	opts = append(opts, sdkmetric.WithReader(reader))

	provider := sdkmetric.NewMeterProvider(opts...)
	otel.SetMeterProvider(provider)

	if err := runtimemetrics.Start(); err != nil {
		return nil, err
	}

	return provider, nil
}

func otlpmetricClient(ctx context.Context, conf *config, dsn *DSN) (sdkmetric.Exporter, error) {
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(dsn.OTLPHttpEndpoint()),
		otlpmetrichttp.WithHeaders(dsn.Headers),
		otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
		otlpmetrichttp.WithTemporalitySelector(preferDeltaTemporalitySelector),
	}

	if conf.tlsConf != nil {
		options = append(options, otlpmetrichttp.WithTLSClientConfig(conf.tlsConf))
	} else if dsn.Scheme == "http" {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	return otlpmetrichttp.New(ctx, options...)
}

func preferDeltaTemporalitySelector(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	switch kind {
	case sdkmetric.InstrumentKindCounter,
		sdkmetric.InstrumentKindObservableCounter,
		sdkmetric.InstrumentKindHistogram:
		return metricdata.DeltaTemporality
	default:
		return metricdata.CumulativeTemporality
	}
}
