package uptrace

import (
	"context"

	"github.com/uptrace/uptrace-go/uptrace"
	"github.com/yoshino-s/go-framework/application"
	"github.com/yoshino-s/go-framework/configuration"
)

type UptraceApp struct {
	*application.EmptyApplication
	config config

	opts []uptrace.Option
}

func NewUptraceApp() *UptraceApp {
	return &UptraceApp{
		EmptyApplication: application.NewEmptyApplication("uptrace"),
	}
}

func (app *UptraceApp) Configuration() configuration.Configuration {
	return &app.config
}

func (app *UptraceApp) SetOptions(opts ...uptrace.Option) {
	app.opts = opts
}

func (app *UptraceApp) BeforeSetup(ctx context.Context) {
	opts := []uptrace.Option{}
	if app.config.DSN != "" {
		opts = append(opts, uptrace.WithDSN(app.config.DSN))
	}

	uptrace.ConfigureOpentelemetry(opts...)
}

func (app *UptraceApp) Close(ctx context.Context) {
	uptrace.Shutdown(ctx)
}
