package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func (app *OtlpApp) configureLogging(ctx context.Context) (*sdklog.LoggerProvider, error) {
	var opts []sdklog.LoggerProviderOption
	opts = append(opts, sdklog.WithResource(app.resource))

	dsn, err := ParseDSN(app.config.dsn)
	if err != nil {
		return nil, err
	}

	exp, err := newOtlpLogExporter(ctx, app.config, dsn)
	if err != nil {
		return nil, err
	}

	queueSize := queueSize()
	bspOptions := []sdklog.BatchProcessorOption{
		sdklog.WithMaxQueueSize(queueSize),
		sdklog.WithExportMaxBatchSize(queueSize),
		sdklog.WithExportInterval(10 * time.Second),
		sdklog.WithExportTimeout(10 * time.Second),
	}
	bsp := sdklog.NewBatchProcessor(exp, bspOptions...)
	opts = append(opts, sdklog.WithProcessor(bsp))

	provider := sdklog.NewLoggerProvider(opts...)
	global.SetLoggerProvider(provider)

	return provider, nil
}

func newOtlpLogExporter(
	ctx context.Context, conf *config, dsn *DSN,
) (*otlploghttp.Exporter, error) {
	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(dsn.OTLPHttpEndpoint()),
		otlploghttp.WithHeaders(dsn.Headers),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	}

	if conf.tlsConf != nil {
		options = append(options, otlploghttp.WithTLSClientConfig(conf.tlsConf))
	} else if dsn.Scheme == "http" {
		options = append(options, otlploghttp.WithInsecure())
	}

	return otlploghttp.New(ctx, options...)
}
