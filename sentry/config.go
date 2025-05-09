package sentry

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoshino-s/go-framework/configuration"
	"github.com/yoshino-s/go-framework/utils"
)

var _ configuration.Configuration = (*telemetryConfiguration)(nil)

type telemetryConfiguration struct {
	SentryDSN        string  `mapstructure:"sentry_dsn"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
}

func (t *telemetryConfiguration) Register(flagSet *pflag.FlagSet) {
	flagSet.String("telemetry.sentry_dsn", "", "sentry dsn")
	flagSet.Float64("telemetry.traces_sample_rate", 1.0, "traces sample rate")
	utils.MustNoError(viper.BindPFlags(flagSet))
	configuration.Register(t)
}

func (c *telemetryConfiguration) Read() {
	utils.MustDecodeFromMapstructure(viper.AllSettings()["telemetry"], c)
}
