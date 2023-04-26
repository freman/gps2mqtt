package protocol

import (
	"github.com/freman/gps2mqtt/mqtt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Configerer interface {
	ProtocolConfiguration(name string, v interface{}) (err error)
	Whitelist(name string) bool
}

type Interface interface {
	Setup(config Configerer) error
	Run(chan mqtt.Identifier) error
}

var interfaces = map[string]func(zerolog.Logger) Interface{}

func Register(name string, f func(zerolog.Logger) Interface) {
	interfaces[name] = f
}

func Get(name string) Interface {
	if ifacef, ok := interfaces[name]; ok {
		return ifacef(log.With().Str("protocol", name).Logger())
	}

	return nil
}
