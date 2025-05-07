package fofa

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoshino-s/go-framework/configuration"
	"github.com/yoshino-s/go-framework/utils"
)

var _ configuration.Configuration = &config{}

type config struct {
	Email    string
	Key      string
	Endpoint string
}

func (c *config) Register(set *pflag.FlagSet) {
	set.String("fofa.email", "", "fofa email")
	set.String("fofa.key", "", "fofa key")
	set.String("fofa.endpoint", "https://fofa.info/api/v1", "fofa endpoint")

	utils.MustNoError(viper.BindPFlags(set))
	configuration.Register(c)
}

func (c *config) Read() {
	utils.MustDecodeFromMapstructure(viper.AllSettings()["fofa"], c)
}
