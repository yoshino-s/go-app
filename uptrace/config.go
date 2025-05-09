package uptrace

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoshino-s/go-framework/configuration"
	"github.com/yoshino-s/go-framework/utils"
)

var _ configuration.Configuration = (*config)(nil)

type config struct {
	DSN string `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
}

func (c *config) Register(set *pflag.FlagSet) {
	set.String("uptrace.dsn", "", "uptrace dsn")
	utils.MustNoError(viper.BindPFlags(set))

	configuration.Register(c)
}

func (c *config) Read() {
	utils.MustDecodeFromMapstructure(viper.AllSettings()["uptrace"], c)
}
