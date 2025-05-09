package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/yoshino-s/go-framework/application"
	"github.com/yoshino-s/go-framework/common"
	"github.com/yoshino-s/go-framework/configuration"
	"go.uber.org/zap"
)

var _ application.Application = &Telemetry{}

type Telemetry struct {
	*application.EmptyApplication
	config telemetryConfiguration
}

func New() *Telemetry {
	return &Telemetry{
		EmptyApplication: application.NewEmptyApplication("telemetry"),
	}
}

func (t *Telemetry) Configuration() configuration.Configuration {
	return &t.config
}

func (t *Telemetry) BeforeSetup(context context.Context) {
	if t.config.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              t.config.SentryDSN,
			Debug:            common.IsDev(),
			EnableTracing:    true,
			AttachStacktrace: true,
			TracesSampleRate: t.config.TracesSampleRate,
			SendDefaultPII:   true,
			Release:          common.Version,
		})
		if err != nil {
			t.Logger.Error("sentry.Init failed", zap.Error(err))
		}
	}
}

func IsSentryInitialized() bool {
	return sentry.CurrentHub().Client() != nil
}
