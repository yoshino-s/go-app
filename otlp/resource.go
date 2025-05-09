package otlp

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
)

func (app *OtlpApp) newResource(ctx context.Context) *resource.Resource {
	if app.config.resource != nil {
		if len(app.config.resourceAttributes) > 0 {
			app.Logger.Sugar().Warnf("WithResource overrides WithResourceAttributes (discarding %v)",
				app.config.resourceAttributes)
		}
		if len(app.config.resourceDetectors) > 0 {
			app.Logger.Sugar().Warnf("WithResource overrides WithResourceDetectors (discarding %v)",
				app.config.resourceDetectors)
		}
		return app.config.resource
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithDetectors(app.config.resourceDetectors...),
		resource.WithAttributes(app.config.resourceAttributes...))
	if err != nil {
		otel.Handle(err)
		return resource.Environment()
	}
	return res
}
