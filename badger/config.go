package badger

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoshino-s/go-framework/configuration"
	"github.com/yoshino-s/go-framework/utils"
)

var _ configuration.Configuration = (*config)(nil)

type config struct {
	Path string `mapstructure:"path"`
}

func (c *config) Register(flagSet *pflag.FlagSet) {
	flagSet.String("badger.path", "/tmp/badger-db", "The path of badger store")
	utils.MustNoError(viper.BindPFlags(flagSet))
	configuration.Register(c)
}

func (c *config) Read() {
	utils.MustDecodeFromMapstructure(viper.AllSettings()["badger"], c)
}
