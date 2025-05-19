package fofa

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/pflag"
	"github.com/yoshino-s/go-framework/configuration"
)

func TestFofaApp(t *testing.T) {
	app := New()
	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	app.Configuration().Register(flagSet)

	configuration.Setup("test")

	Convey("Setup", t, func() {
		app.Setup(t.Context())
	})

	Convey("FofaApp", t, func() {
		res, err := app.Query(t.Context(), "ip=8.8.8.8", 1, 10)
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 10)
		fmt.Println(res)
	})
}
