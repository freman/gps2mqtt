package main

import (
	"encoding/json"
	"flag"
	"os"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/freman/gps2mqtt"
	"github.com/freman/gps2mqtt/homeassistant"
	"github.com/freman/gps2mqtt/mqtt"
	"github.com/freman/gps2mqtt/protocol"
	_ "github.com/freman/gps2mqtt/protocol/watch"
)

func main() {
	pCfg := flag.String("config", "config.toml", "Path to the configuration file")
	pHuman := flag.Bool("pretty", false, "Pretty human readable log output")

	flag.Parse()

	cfg, err := gps2mqtt.LoadConfiguration(*pCfg)
	if err != nil {
		panic(err)
	}

	if *pHuman {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	chMessage := make(chan mqtt.Message, 10)

	opts := paho.NewClientOptions().
		SetClientID(cfg.MQTT.ClientName).
		SetKeepAlive(cfg.MQTT.KeepAlive).
		SetPingTimeout(cfg.MQTT.PingTimeout)

	if cfg.MQTT.Username != "" && cfg.MQTT.Password != "" {
		opts.SetUsername(cfg.MQTT.Username)
		opts.SetPassword(cfg.MQTT.Password)
	}

	for _, broker := range cfg.MQTT.Brokers {
		opts.AddBroker(broker)
	}

	opts.SetWill("gps2mqtt/availability", "offline", 0, false)

	c := paho.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("Failed to connect to MQTT Broker.")
	}

	for i := range cfg.Protocols {
		p := protocol.Get(i)
		if p == nil {
			log.Fatal().Str("protocol", i).Msg("Unsupported GPS protocol specified in configuration.")
		}

		if err := p.Setup(cfg); err != nil {
			log.Fatal().Str("protocol", i).Err(err).Msg("GPS protocol configuration failed.")
		}

		go p.Run(chMessage)
	}

	c.Publish("gps2mqtt/availability", 0, false, "online") // TODO error check

	for msg := range chMessage {
		mqttid := msg.Data.MQTTID()
		topicPrefix := "gps2mqtt/device/" + mqttid
		deviceID := msg.Data.Device()

		switch msg.Type {
		case mqtt.TypeHello:
			meta, has := cfg.Meta[deviceID]
			if has && meta.Name != "" {
				hc := homeassistant.AutoConfiguration{
					Name:                meta.Name,
					Icon:                meta.Icon,
					StateTopic:          topicPrefix,
					AvailabilityTopic:   "gps2mqtt/availability",
					JSONAttributesTopic: topicPrefix + "/attributes",
					SourceType:          "gps",
				}

				b, err := json.Marshal(hc)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to marshal configuration message.")
				}

				c.Publish("homeassistant/device_tracker/"+mqttid+"/config", 0, false, b) // TODO error check
			}
		case mqtt.TypeUpdate:
			b, err := json.Marshal(msg.Data)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to marshal update message.")
			}

			c.Publish(topicPrefix+"/attributes", 0, false, b) // TODO error check
		}
	}
}