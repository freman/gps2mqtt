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

func (p *Parser) ReadPacket() (packet interface{}, err error) {
	var char byte
	if char, err = p.reader.ReadByte(); err != nil {
		return packet, err
	}

	var np Packet
	if char == '[' {
		np.Company, err = p.readString('*')
		if err != nil {
			return
		}

		np.DeviceID, err = p.readString('*')
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

		np.ContentLength = binary.BigEndian.Uint16(hlen)

		if np.Content, err = p.readString(']'); err != nil {
			return
		}

		return p.MutatePacket(np)
	}

	return nil, errors.New("bad packet")
}

func (p Parser) MutatePacket(packet Packet) (interface{}, error) {
	var err error

	if packet.ContentLength == 0 {
		return nil, errors.New("invalid packet")
	}

	if packet.Content[0:2] == "LK" {
		return &PacketLK{&packet}, nil
	}

	if packet.Content[0:2] == "UD" {
		content := strings.Split(packet.Content, ",")

		np := &PacketUD{
			Packet: &packet,
		}

		if np.Timestamp, err = time.Parse("020106150405", content[1]+content[2]); err != nil {
			return nil, err
		}

		np.Position = strings.EqualFold(content[3], "A")

		if np.Latitude, err = strconv.ParseFloat(content[4], 64); err != nil {
			return nil, err
		}

		if strings.EqualFold(content[5], "S") {
			np.Latitude *= -1.0
		}

		if np.Longitude, err = strconv.ParseFloat(content[6], 64); err != nil {
			return nil, err
		}

		if strings.EqualFold(content[7], "W") {
			np.Longitude *= -1.0
		}

		if np.Speed, err = strconv.ParseFloat(content[8], 64); err != nil {
			return nil, err
		}

		np.Speed *= 1.60934

		if np.Heading, err = strconv.ParseFloat(content[9], 64); err != nil {
			return nil, err
		}

		if np.Altitude, err = strconv.ParseFloat(content[10], 64); err != nil {
			return nil, err
		}

		if np.Satellites, err = strconv.ParseInt(content[11], 10, 64); err != nil {
			return nil, err
		}

		if np.RSSI, err = strconv.ParseFloat(content[12], 64); err != nil {
			return nil, err
		}

		if np.Battery, err = strconv.ParseFloat(content[13], 64); err != nil {
			return nil, err
		}

		return np, nil
	}

	return p, nil
}
