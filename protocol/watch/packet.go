package watch

import (
	"fmt"
	"io"
	"time"
)

type Packet struct {
	Company       string `json:"company"`
	DeviceID      string `json:"device_id"`
	ContentLength uint16 `json:"content_length"`
	Content       string `json:"content,omitempty"`

	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Position  bool      `json:"position"`

	Satellites int64   `json:"satellites"`
	RSSI       float64 `json:"rssi"`
	Battery    float64 `json:"battery"`

	packetType string
}

func (p *Packet) MQTTID() string {
	return fmt.Sprintf("%s_%s", p.Company, p.DeviceID)
}

func (p *Packet) Device() string {
	return fmt.Sprintf("%s*%s", p.Company, p.DeviceID)
}

func (p *Packet) Respond(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `[%s*%s*0002*%s]`, p.Company, p.DeviceID, p.packetType)
	return err
}

func (p *Packet) WantResponse() bool {
	return p.packetType == "LK"
}

func (p *Packet) Valid() bool {
	return p.packetType != "LK"
}
