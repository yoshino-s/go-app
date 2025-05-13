package telemetry

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type OtlpApp struct {
	config *config

	dsn    *DSN
	tracer trace.Tracer

	TraceProvider  *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider

	resource *resource.Resource
}

func New(ctx context.Context, opts ...Option) *OtlpApp {
	app := &OtlpApp{}

	conf := newConfig(opts)
	app.config = conf

	atomicClient.Store(app)

	if conf.dsn != "" {
		app.setup(ctx)
	}

	return app
}

func (app *OtlpApp) setup(ctx context.Context) {
	dsn, err := ParseDSN(app.config.dsn)
	if err != nil {
		app.config.logger.Sugar().Panicf("invalid Uptrace DSN: %s (Uptrace is disabled)", err)
		return
	}
	app.dsn = dsn
	app.resource = app.newResource(ctx)

	configurePropagator(app.config)

	if app.config.tracingEnabled {
		tp, err := app.configureTracing(ctx)
		if err != nil {
			app.config.logger.Error("configureTracing failed", zap.Error(err))
		}
		app.TraceProvider = tp
	}
	if app.config.metricsEnabled {
		mp, err := app.configureMetrics(ctx)
		if err != nil {
			app.config.logger.Error("configureMetrics failed", zap.Error(err))
		}
		app.MeterProvider = mp
	}
	if app.config.loggingEnabled {
		lp, err := app.configureLogging(ctx)
		if err != nil {
			app.config.logger.Error("configureLogging failed", zap.Error(err))
		}
		app.LoggerProvider = lp
	}
}

func configurePropagator(conf *config) {
	textMapPropagator := conf.textMapPropagator
	if textMapPropagator == nil {
		textMapPropagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}
	otel.SetTextMapPropagator(textMapPropagator)
}

const dummySpanName = "__dummy__"

func (c *OtlpApp) Close(ctx context.Context) {
	if c.TraceProvider != nil {
		if err := c.TraceProvider.Shutdown(ctx); err != nil {
			c.config.logger.Error("TraceProvider Shutdown failed", zap.Error(err))
		}
		c.TraceProvider = nil
	}
	if c.MeterProvider != nil {
		if err := c.MeterProvider.Shutdown(ctx); err != nil {
			c.config.logger.Error("MeterProvider Shutdown failed", zap.Error(err))
		}
		c.MeterProvider = nil
	}
	if c.LoggerProvider != nil {
		if err := c.LoggerProvider.Shutdown(ctx); err != nil {
			c.config.logger.Error("LoggerProvider Shutdown failed", zap.Error(err))
		}
		c.LoggerProvider = nil
	}
}

func (c *OtlpApp) ForceFlush(ctx context.Context) (lastErr error) {
	if c.TraceProvider != nil {
		if err := c.TraceProvider.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	if c.MeterProvider != nil {
		if err := c.MeterProvider.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	if c.LoggerProvider != nil {
		if err := c.LoggerProvider.ForceFlush(ctx); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// TraceURL returns the trace URL for the span.
func (c *OtlpApp) TraceURL(span trace.Span) string {
	sctx := span.SpanContext()
	return fmt.Sprintf("%s/traces/%s?span_id=%s",
		c.dsn.SiteURL(), sctx.TraceID(), sctx.SpanID().String())
}

// ReportError reports an error as a span event creating a dummy span if necessary.
func (c *OtlpApp) ReportError(ctx context.Context, err error, opts ...trace.EventOption) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		_, span = c.tracer.Start(ctx, dummySpanName)
		defer span.End()
	}

	span.RecordError(err, opts...)
}

// ReportPanic is used with defer to report panics.
func (c *OtlpApp) ReportPanic(ctx context.Context, val any) {
	c.reportPanic(ctx, val)
	// Force flush since we are about to exit on panic.
	if c.TraceProvider != nil {
		_ = c.TraceProvider.ForceFlush(ctx)
	}
}

func (c *OtlpApp) reportPanic(ctx context.Context, val interface{}) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		_, span = c.tracer.Start(ctx, dummySpanName)
		defer span.End()
	}

	stackTrace := make([]byte, 2048)
	n := runtime.Stack(stackTrace, false)

	span.AddEvent(
		"log",
		trace.WithAttributes(
			attribute.String("log.severity", "panic"),
			attribute.String("log.message", fmt.Sprint(val)),
			attribute.String("exception.stackstrace", string(stackTrace[:n])),
		),
	)
}

var (
	atomicClient atomic.Value
)

func ActiveClient() *OtlpApp {
	v := atomicClient.Load()
	if v == nil {
		return nil
	}
	return v.(*OtlpApp)
}

// ReportError is a helper function to report an error as a span event.
// If the current span is not recording, it creates a new dummy span and ends it immediately.
// It returns the error to allow for further handling.
func ReportError(ctx context.Context, err error, opts ...trace.EventOption) error {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		_, span = ActiveClient().tracer.Start(ctx, dummySpanName)
		defer span.End()
	}
	span.RecordError(err, opts...)
	return err
}

func ForceFlush(ctx context.Context) error {
	return ActiveClient().ForceFlush(ctx)
}
