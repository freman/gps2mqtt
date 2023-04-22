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
}

type PacketLK struct {
	*Packet
}

type PacketUD struct {
	*Packet
	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Altitude  float64   `json:"altitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Position  bool      `json:"position"`

	Satelites int64   `json:"satelites"`
	RSSI      float64 `json:"rssi"`
	Battery   float64 `json:"battery"`
}

func (p *Packet) MQTTID() string {
	return fmt.Sprintf("%s_%s", p.Company, p.DeviceID)
}

func (p *Packet) Device() string {
	return fmt.Sprintf("%s*%s", p.Company, p.DeviceID)
}

func (p *PacketLK) Respond(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, `[%s*%s*0002*LK]`, p.Company, p.DeviceID)
	return err
}
