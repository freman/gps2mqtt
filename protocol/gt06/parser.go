package gt06

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/freman/gps2mqtt/checksum"
	"github.com/rs/zerolog"
)

const (
	protoLogin    byte = 0x01
	protoLocation byte = 0x12
	protoStatus   byte = 0x13
	protoString   byte = 0x15
	protoAlarm    byte = 0x16
	protoGPSQuery byte = 0x1A
	protoCommand  byte = 0x80
)

var (
	startMessage = []byte{0x78, 0x78}
	stopMessage  = []byte{0x0D, 0x0A}
)

type Parser struct {
	reader *bufio.Reader
	log    zerolog.Logger

	id hexString
}

func (p *Parser) read(d []byte) error {
	sz, err := p.reader.Read(d)
	if err != nil {
		return err
	}

	if len(d) != sz {
		return errors.New("unexpected number of bytes read")
	}

	return nil
}

func (p *Parser) ReadPacket() (*Packet, error) {
	start := make([]byte, 2)
	stop := make([]byte, 2)

	if err := p.read(start); err != nil {
		return nil, err
	}

	if !bytes.Equal(startMessage, start) {
		return nil, errors.New("unexpected start bits")
	}

	length, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Prepend length to the message now because it's needed for CRC and I'm lazy
	msg := make([]byte, length+1)
	msg[0] = length

	if err := p.read(msg[1:]); err != nil {
		return nil, err
	}

	if err := p.read(stop); err != nil {
		return nil, err
	}

	if !bytes.Equal(stopMessage, stop) {
		return nil, errors.New("unexpected stop bits")
	}

	if err := p.verifyCRC(msg); err != nil {
		return nil, err
	}

	// Strip off the length now cos it causes headaches going forth.
	msg = msg[1:]

	return p.parsePacket(&Packet{
		DeviceID: p.id,
		protocol: msg[0],
		sequence: binary.BigEndian.Uint16(msg[length-4 : length-2]),
	}, msg[1:length-4])
}

func (p *Parser) parsePacket(packet *Packet, msg []byte) (*Packet, error) {
	switch packet.protocol {
	case protoLogin:
		p.id = msg
		packet.DeviceID = msg
		return packet, nil
	case protoStatus:
		// TODO: care about status?
		return packet, nil
	case protoLocation:
		return p.readLocation(packet, bytes.NewReader(msg))
	case protoAlarm:
		// TODO: care about alarm?
		return p.readLocation(packet, bytes.NewReader(msg))
	}

	return nil, errors.New("bad packet")
}

func (p *Parser) readLocation(packet *Packet, reader io.Reader) (*Packet, error) {
	var data packetData

	if err := binary.Read(reader, binary.BigEndian, &data); err != nil {
		return nil, err
	}

	packet.Heading = data.Course()
	packet.Latitude = data.GetLatitude()
	packet.Longitude = data.GetLongitude()
	packet.Position = data.PositionValid()
	packet.Speed = float64(data.Speed)
	packet.Timestamp = data.GetTimestamp()
	packet.Satelites = int(data.GetSatelites())

	return packet, nil
}

func (p *Parser) verifyCRC(msg []byte) error {
	l := len(msg)
	expected := binary.BigEndian.Uint16(msg[l-2:])

	crc := checksum.CRC16_ITU(msg[:l-2])

	if expected != crc {
		return errors.New("invalid crc")
	}

	return nil
}

type packetData struct {
	Year         byte
	Month        byte
	Day          byte
	Hour         byte
	Minute       byte
	Second       byte
	QCSats       byte
	Latitude     uint32
	Longitude    uint32
	Speed        byte
	CourseStatus uint16
}

func (p packetData) GetLatitude() float64 {
	v := float64(p.Latitude) / 30000 / 60
	if p.North() {
		return v
	}

	return -v
}

func (p packetData) GetLongitude() float64 {
	v := float64(p.Longitude) / 30000 / 60
	if p.West() {
		return -v
	}
	return v
}

func (p packetData) ACCIsOn() bool {
	return p.CourseStatus&0x8000 == 0x8000
}

func (p packetData) Input2IsOn() bool {
	return p.CourseStatus&0x4000 == 0x4000
}

func (p packetData) IsDifferential() bool {
	return p.CourseStatus&0x2000 == 0x2000
}

func (p packetData) PositionValid() bool {
	return p.CourseStatus&0x1000 == 0x1000
}

func (p packetData) West() bool {
	return p.CourseStatus&0x800 == 0x800
}

func (p packetData) North() bool {
	return p.CourseStatus&0x400 == 0x400
}

func (p packetData) Course() float64 {
	return float64(p.CourseStatus & 0x3ff)
}

func (p packetData) GetSatelites() byte {
	return p.QCSats & 0x0f
}

func (p packetData) GetTimestamp() time.Time {
	return time.Date(
		time.Now().Year()/100*100+int(p.Year),
		time.Month(p.Month),
		int(p.Day),
		int(p.Hour),
		int(p.Minute),
		int(p.Second),
		0,
		time.UTC,
	)
}
