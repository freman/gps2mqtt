package watch

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	reader *bufio.Reader
}

func (p *Parser) readString(delim byte) (out string, err error) {
	out, err = p.reader.ReadString(delim)
	return strings.TrimSuffix(out, string([]byte{delim})), err
}

func (p *Parser) ReadPacket() (packet *Packet, err error) {
	var char byte
	if char, err = p.reader.ReadByte(); err != nil {
		return packet, err
	}

	if char == '[' {
		packet = &Packet{}

		packet.Company, err = p.readString('*')
		if err != nil {
			return
		}

		packet.DeviceID, err = p.readString('*')
		if err != nil {
			return
		}

		var plen string
		if plen, err = p.readString('*'); err != nil {
			return
		}

		var hlen []byte
		if hlen, err = hex.DecodeString(plen); err != nil {
			return
		}

		packet.ContentLength = binary.BigEndian.Uint16(hlen)

		if packet.Content, err = p.readString(']'); err != nil {
			return
		}

		return p.MutatePacket(packet)
	}

	return nil, errors.New("bad packet")
}

func (p Parser) MutatePacket(packet *Packet) (*Packet, error) {
	var err error

	if packet.ContentLength == 0 {
		return nil, errors.New("invalid packet")
	}

	packet.packetType = packet.Content[0:2]

	if packet.packetType == "UD" {
		content := strings.Split(packet.Content, ",")

		if packet.Timestamp, err = time.Parse("020106150405", content[1]+content[2]); err != nil {
			return nil, err
		}

		packet.Position = strings.EqualFold(content[3], "A")

		if packet.Latitude, err = strconv.ParseFloat(content[4], 64); err != nil {
			return nil, err
		}

		if strings.EqualFold(content[5], "S") {
			packet.Latitude *= -1.0
		}

		if packet.Longitude, err = strconv.ParseFloat(content[6], 64); err != nil {
			return nil, err
		}

		if strings.EqualFold(content[7], "W") {
			packet.Longitude *= -1.0
		}

		if packet.Speed, err = strconv.ParseFloat(content[8], 64); err != nil {
			return nil, err
		}

		packet.Speed *= 1.60934

		if packet.Heading, err = strconv.ParseFloat(content[9], 64); err != nil {
			return nil, err
		}

		if packet.Altitude, err = strconv.ParseFloat(content[10], 64); err != nil {
			return nil, err
		}

		if packet.Satellites, err = strconv.ParseInt(content[11], 10, 64); err != nil {
			return nil, err
		}

		if packet.RSSI, err = strconv.ParseFloat(content[12], 64); err != nil {
			return nil, err
		}

		if packet.Battery, err = strconv.ParseFloat(content[13], 64); err != nil {
			return nil, err
		}
	}

	return packet, nil
}
