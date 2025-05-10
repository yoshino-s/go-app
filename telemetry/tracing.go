package telemetry

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"log/slog"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func (app *OtlpApp) configureTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
	provider := app.config.tracerProvider
	if provider == nil {
		var opts []sdktrace.TracerProviderOption

		opts = append(opts, sdktrace.WithIDGenerator(newIDGenerator()))
		opts = append(opts, sdktrace.WithResource(app.resource))
		if app.config.traceSampler != nil {
			opts = append(opts, sdktrace.WithSampler(app.config.traceSampler))
		}

		provider = sdktrace.NewTracerProvider(opts...)
		otel.SetTracerProvider(provider)
	}

	dsn, err := ParseDSN(app.config.dsn)
	if err != nil {
		return nil, err
	}

	exp, err := otlptrace.New(ctx, otlpTraceClient(app.config, dsn))
	if err != nil {
		return nil, err
	}

	queueSize := queueSize()
	bspOptions := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithMaxQueueSize(queueSize),
		sdktrace.WithMaxExportBatchSize(queueSize),
		sdktrace.WithBatchTimeout(10 * time.Second),
		sdktrace.WithExportTimeout(10 * time.Second),
	}
	bspOptions = append(bspOptions, app.config.bspOptions...)

	bsp := sdktrace.NewBatchSpanProcessor(exp, bspOptions...)
	provider.RegisterSpanProcessor(bsp)

	// Register additional span processors.
	for _, sp := range app.config.spanProcessors {
		provider.RegisterSpanProcessor(sp)
	}

	if app.config.prettyPrint {
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			slog.Error("stdouttrace.New failed", slog.Any("err", err))
		} else {
			provider.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(exporter))
		}
	}

	return provider, nil
}

func otlpTraceClient(conf *config, dsn *DSN) otlptrace.Client {
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(dsn.OTLPHttpEndpoint()),
		otlptracehttp.WithHeaders(dsn.Headers),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	}

	if conf.tlsConf != nil {
		options = append(options, otlptracehttp.WithTLSClientConfig(conf.tlsConf))
	} else if dsn.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.NewClient(options...)
}

func queueSize() int {
	const min = 1000
	const max = 16000

	n := (runtime.GOMAXPROCS(0) / 2) * 1000
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

//------------------------------------------------------------------------------

const spanIDPrec = int64(time.Millisecond)

type idGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func newIDGenerator() *idGenerator {
	gen := &idGenerator{}
	var rngSeed int64
	_ = binary.Read(cryptorand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}

var _ sdktrace.IDGenerator = (*idGenerator)(nil)

// NewIDs returns a new trace and span ID.
func (gen *idGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	unixNano := time.Now().UnixNano()

	gen.Lock()
	defer gen.Unlock()

	tid := trace.TraceID{}
	binary.BigEndian.PutUint64(tid[:8], uint64(unixNano))
	_, _ = gen.randSource.Read(tid[8:])

	sid := trace.SpanID{}
	binary.BigEndian.PutUint32(sid[:4], uint32(unixNano/spanIDPrec))
	_, _ = gen.randSource.Read(sid[4:])

	return tid, sid
}

// NewSpanID returns a ID for a new span in the trace with traceID.
func (gen *idGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	unixNano := time.Now().UnixNano()

	gen.Lock()
	defer gen.Unlock()

	sid := trace.SpanID{}
	binary.BigEndian.PutUint32(sid[:4], uint32(unixNano/spanIDPrec))
	_, _ = gen.randSource.Read(sid[4:])

	return sid
}
