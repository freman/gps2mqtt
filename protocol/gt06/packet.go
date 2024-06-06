package gt06

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/freman/gps2mqtt/checksum"
)

type Packet struct {
	DeviceID  hexString `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	Position  bool      `json:"position"`
	Satelites int       `json:"satelites"`

	protocol byte
	sequence uint16
}

func (p *Packet) MQTTID() string {
	return p.DeviceID.String()
}

func (p *Packet) Device() string {
	return p.DeviceID.String()
}

func (p *Packet) Respond(writer io.Writer) (err error) {
	var buf bytes.Buffer
	buf.Write(startMessage)
	buf.WriteByte(0x05)
	buf.WriteByte(p.protocol)

	if err := binary.Write(&buf, binary.BigEndian, p.sequence); err != nil {
		return err
	}

	buf.Write([]byte{0, 0})
	buf.Write(stopMessage)

	b := buf.Bytes()
	binary.BigEndian.PutUint16(b[6:8], checksum.CRC16_ITU(b[1:6]))

	_, err = writer.Write(b)

	return err
}

func (p *Packet) WantsResponse() bool {
	return true
}

func (p *Packet) Valid() bool {
	return p.protocol == protoLocation
}

type hexString []byte

func (h *hexString) String() string {
	return strings.TrimPrefix(fmt.Sprintf("%x", *h), "0")
}

func (h *hexString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}
