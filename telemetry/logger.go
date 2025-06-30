package telemetry

import (
	"github.com/yoshino-s/go-framework/application"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.uber.org/zap/zapcore"
)

var _ application.LoggerCoreContrib = (*OtlpApp)(nil)

func (o *OtlpApp) LoggerCores() []zapcore.Core {
	return []zapcore.Core{
		otelzap.NewCore("otelzap"),
	}
}
