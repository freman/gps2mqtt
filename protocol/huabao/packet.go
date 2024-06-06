package huabao

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/freman/gps2mqtt/checksum"
)

const (
	protoRegister         uint16 = 0x0100
	protoRegisterResponse uint16 = 0x8100
	protoTerminalAuth     uint16 = 0x0102
	protoGeneralResponse  uint16 = 0x8001
	protoHeartbeat        uint16 = 0x0002
	protoLocationReport   uint16 = 0x0200
)

type properties uint16

type statusFlags uint32

type header struct {
	MessageType uint16
	Properties  properties
	Terminal    terminalBCD
	Sequence    uint16
}

func (p properties) Length() int {
	return int(p & (2<<9 - 1))
}

func (p properties) Encrypted() bool {
	return p>>10&3 > 0
}

func (p properties) SubPackage() bool {
	return p>>13&1 == 1
}

// ACC is on if true
func (s statusFlags) ACC() bool {
	return s&1 == 1
}

// Positioning is on if true
func (s statusFlags) Positioning() bool {
	return s&2 == 2
}

// Latitude is south if true
func (s statusFlags) Latitude() bool {
	return s&4 == 4
}

// Longitude is west if true
func (s statusFlags) Longitude() bool {
	return s&8 == 8
}

type basicLocationInformation struct {
	AlarmFlags uint32
	Status     statusFlags
	Latitude   uint32
	Longitude  uint32
	Altitude   uint16
	Speed      uint16
	Heading    uint16
	Timestamp  terminalBCD
}

type Packet struct {
	DeviceID string `json:"device_id"`

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

	ManufacturerID string `json:"manufacturer"`
	TerminalModel  string `json:"model"`
	TerminalID     string `json:"terminal_id"`

	header header
}

func (p *Packet) MQTTID() string {
	return p.DeviceID
}

func (p *Packet) Device() string {
	return p.DeviceID
}

func (p *Packet) Respond(wr io.Writer) error {
	switch p.header.MessageType {
	case protoRegister:
		return p.respondToRegister(wr)
	case protoTerminalAuth, protoHeartbeat, protoLocationReport:
		return p.sendGeneralResponse(wr)
	}

	return errors.New("unknown message type")
}

func (p *Packet) respondToRegister(wr io.Writer) error {
	var body bytes.Buffer
	binary.Write(&body, binary.BigEndian, p.header.Sequence)
	body.WriteByte(0)
	body.WriteString(p.DeviceID)

	return p.respondWith(protoRegisterResponse, body, wr)
}

func (p *Packet) sendGeneralResponse(wr io.Writer) error {
	var body bytes.Buffer
	binary.Write(&body, binary.BigEndian, p.header.Sequence)
	binary.Write(&body, binary.BigEndian, p.header.MessageType)
	body.WriteByte(0)

	return p.respondWith(protoGeneralResponse, body, wr)
}

func (p *Packet) respondWith(messageType uint16, body bytes.Buffer, wr io.Writer) error {
	var buf bytes.Buffer
	buf.WriteByte(0x7e)

	binary.Write(&buf, binary.BigEndian, header{
		MessageType: messageType,
		Properties:  properties(body.Len()),
		Terminal:    p.header.Terminal,
		Sequence:    0,
	})

	buf.Write(body.Bytes())

	buf.WriteByte(checksum.XOR(buf.Bytes()[1:]))
	buf.WriteByte(0x7e)

	_, err := writer{wr}.Write(buf.Bytes())
	return err
}

func (p *Packet) WantResponse() bool {
	return true
}

func (p *Packet) Valid() bool {
	return p.header.MessageType == protoLocationReport
}

func (p *Packet) importBasicInformation(rep basicLocationInformation) {
	p.Altitude = float64(rep.Altitude)
	p.Heading = float64(rep.Heading)
	p.Latitude = float64(rep.Latitude) * 0.000001
	p.Longitude = float64(rep.Longitude) * 0.000001
	p.Speed = float64(rep.Speed) * 0.1

	var err error
	p.Timestamp, err = time.ParseInLocation("060102150405", rep.Timestamp.String(), time.FixedZone("GMT+8", 8*60*60))
	if err != nil {
		fmt.Println(err)
	}

	if rep.Status.Latitude() {
		p.Latitude *= -1
	}

	if rep.Status.Longitude() {
		p.Longitude *= -1
	}

	p.Position = rep.Status.Positioning()
}

type terminalBCD [6]byte

func (h *terminalBCD) String() string {
	return fmt.Sprintf("%x", *h)
}

func (h *terminalBCD) MarshalJSON() ([]byte, error) {
	return []byte(`"` + h.String() + `"`), nil
}
