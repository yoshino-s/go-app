package otlp

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestOtlp(t *testing.T) {
	logger := otelzap.New(zaptest.NewLogger(t), otelzap.WithMinLevel(zap.DebugLevel))
	Convey("Test Otlp", t, func() {
		app := New(
			WithDSN("https://signoz-otl-http.yoshino-s.xyz/"),
			WithServiceName("test"),
		)
		app.Setup(t.Context())

		fmt.Println(app.dsn)
		logger.Ctx(t.Context()).Info("test")
		fmt.Println(global.GetLoggerProvider())

		app.Close(t.Context())
	})
}
