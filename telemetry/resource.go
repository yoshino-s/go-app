package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
)

func (app *OtlpApp) newResource(ctx context.Context) *resource.Resource {
	if app.config.resource != nil {
		if len(app.config.resourceAttributes) > 0 {
			app.config.logger.Warn("WithResource overrides WithResourceAttributes", zap.Any("attributes", app.config.resourceAttributes))
		}
		if len(app.config.resourceDetectors) > 0 {
			app.config.logger.Warn("WithResource overrides WithResourceDetectors", zap.Any("detectors", app.config.resourceDetectors))
		}
		return app.config.resource
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(app.config.resourceDetectors...),
		resource.WithAttributes(app.config.resourceAttributes...))
	if err != nil {
		otel.Handle(err)
		return resource.Environment()
	}
	return res
}
