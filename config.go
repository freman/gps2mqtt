package gps2mqtt

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	md toml.MetaData `toml:"-"`

	MQTT      ConfigMQTT
	Status    ConfigStatus
	Meta      map[string]ConfigMeta
	Protocols map[string]toml.Primitive `toml:"protocol"`
}

type ConfigMQTT struct {
	ClientName  string
	Brokers     []string
	KeepAlive   time.Duration
	PingTimeout time.Duration
	Username    string
	Password    string
}

type ConfigStatus struct {
	Enabled bool
	Listen  string
}

type ConfigMeta struct {
	Name string
	Icon string
}

func LoadConfiguration(file string) (*Config, error) {
	config := Config{
		MQTT: ConfigMQTT{
			ClientName:  "gps2mqtt",
			KeepAlive:   time.Minute,
			PingTimeout: time.Second,
			Username:    os.Getenv("MQTT_USERNAME"),
			Password:    os.Getenv("MQTT_PASSWORD"),
		},
		Status: ConfigStatus{
			Enabled: false,
			Listen:  "127.0.0.1:8080",
		},
	}

	var err error
	if config.md, err = toml.DecodeFile(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) ProtocolConfiguration(name string, v interface{}) (err error) {
	if c.md.IsDefined("protocol", name) {
		err = c.md.PrimitiveDecode(c.Protocols[name], v)
		return
	}

	return fmt.Errorf("unable to find %s in configuration under protocol", name)
}

func (c *Config) Whitelist(name string) bool {
	_, exists := c.Meta[name]
	return exists
}
