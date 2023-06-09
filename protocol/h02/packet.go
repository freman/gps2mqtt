package h02

import (
	"fmt"
	"io"
	"time"
)

type Packet struct {
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Position  bool      `json:"position"`
	Battery   float64   `json:"battery"`

	packetType string
}

func (p *Packet) MQTTID() string {
	return p.DeviceID
}

func (p *Packet) Device() string {
	return p.DeviceID
}

func (p *Packet) Respond(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `*HQ,%s,V4,V1,%s#`, p.DeviceID, time.Now().In(time.UTC).Format(`20060102150405`))
	return err
}

func (p *Packet) WantsResponse() bool {
	return p.packetType == "HQ:V1"
}

func (p *Packet) Valid() bool {
	return true
}
