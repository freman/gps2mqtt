package h02

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Parser struct {
	reader *bufio.Reader
	log    zerolog.Logger
}

type errUnsupportedPacket struct {
	packetType string
}

func (e errUnsupportedPacket) Is(err error) bool {
	_, isa := err.(*errUnsupportedPacket)
	return isa
}
func (e errUnsupportedPacket) Error() string {
	return fmt.Sprintf("unable to parse %q is an unsupported packet type", e.packetType)
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

	switch char {
	case '*':
		packet, err := p.readAsciiPacket()
		if err != nil {
			return nil, fmt.Errorf("ascii error: %w", err)
		}

		return packet, nil
	case '$':
		packet, err := p.readBinaryPacket()
		if err != nil {
			return nil, fmt.Errorf("binary error: %w", err)
		}

		return packet, nil
	}

	return nil, errors.New("bad packet")
}

func (p *Parser) readAsciiPacket() (packet *Packet, err error) {
	var str string

	if str, err = p.readString('#'); err != nil {
		return packet, err
	}

	data := strings.Split(str, ",")
	switch data[0] + ":" + data[2] {
	case "HQ:V1": // Location
		packet = &Packet{
			packetType: "HQ:V1",

			DeviceID: data[1],
			Position: strings.EqualFold(data[4], "A"),
		}

		ts := data[11] + data[3]
		if packet.Timestamp, err = time.Parse("020106150405", ts); err != nil {
			return nil, fmt.Errorf("%w (%s != 020106150405)", err, ts)
		}

		if packet.Latitude, err = strLatitude(data[5], strings.EqualFold(data[6], "S")); err != nil {
			return nil, fmt.Errorf("failed to parse latitude (%s): %w", data[6], err)
		}

		if packet.Longitude, err = strLongitude(data[7], strings.EqualFold(data[8], "W")); err != nil {
			return nil, fmt.Errorf("failed to parse longitude (%s): %w", data[8], err)
		}

		if packet.Speed, err = strconv.ParseFloat(data[9], 64); err != nil {
			return nil, fmt.Errorf("failed to parse speed (%s): %w", data[9], err)
		}

		packet.Speed *= 1.852 // Convert from knots to km/hr

		if data[10] != "" { // Null is 0 apparently
			if packet.Heading, err = strconv.ParseFloat(data[10], 64); err != nil {
				return nil, fmt.Errorf("failed to parse heading (%s): %w", data[10], err)
			}
		}

		if data[17] != "" {
			tmpi, err := strconv.Atoi(data[17])
			if err != nil {
				return nil, fmt.Errorf("failed to parse battery (%s): %w", data[17], err)
			}

			packet.Battery = batteryConversion(tmpi)
		}

		return packet, nil
	case "HQ:V19": // Sim data
		return nil, &errUnsupportedPacket{"HQ:V19"}
	}

	return nil, errors.New("bad packet (" + data[0] + ":" + data[2] + ")")
}

func (p *Parser) readBinaryPacket() (packet *Packet, err error) {
	data := make([]byte, 50)

	c, err := p.reader.Read(data)
	if err != nil {
		return nil, err
	}

	if c < 50 {
		return nil, errors.New("not enough data")
	}

	packet = &Packet{
		packetType: "$",
		DeviceID:   hex.EncodeToString(data[0:5]),
		Battery:    batteryConversion(int(data[15])),
		Position:   data[20]&2 == 2,
	}

	ts := hex.EncodeToString(data[5:11])
	if packet.Timestamp, err = time.Parse("150405060102", ts); err != nil {
		return nil, fmt.Errorf("%w (%s != 150405060102)", err, ts)
	}

	west := data[20]&8 == 8
	south := data[20]&4 == 4
	hexLongitude := hex.EncodeToString(data[16:21])

	if packet.Longitude, err = strLongitude(hexLongitude[0:5]+"."+hexLongitude[5:9], south); err != nil {
		return nil, fmt.Errorf("failed to parse longitude (%s): %w", hexLongitude[0:5]+"."+hexLongitude[5:9], err)
	}

	if packet.Latitude, err = strLatitude(hex.EncodeToString(data[11:13])+"."+hex.EncodeToString(data[13:15]), west); err != nil {
		return nil, fmt.Errorf("failed to parse latitude (%s): %w", hex.EncodeToString(data[11:13])+"."+hex.EncodeToString(data[13:15]), err)
	}

	hexSpeedDir := hex.EncodeToString(data[21:24])
	if packet.Speed, err = strconv.ParseFloat(hexSpeedDir[0:3], 64); err != nil {
		return nil, fmt.Errorf("failed to parse speed (%s): %w", hexSpeedDir[0:3], err)
	}

	packet.Speed *= 1.852 // Convert from knots to km/hr

	if packet.Heading, err = strconv.ParseFloat(hexSpeedDir[3:6], 64); err != nil {
		return nil, fmt.Errorf("failed to parse heading (%s): %w", hexSpeedDir[3:6], err)
	}

	return packet, nil
}

func strLongitude(pos string, east bool) (float64, error) {
	return strDDMtoDD(3, pos, east)
}

func strLatitude(pos string, south bool) (float64, error) {
	return strDDMtoDD(2, pos, south)
}

// strDDMtoDD will convert the string representation of Decimal Degrees and Minutes to
// Decimal Degrees, x is 2 for Latitude or 3 for Longitude.
func strDDMtoDD(x int, pos string, dir bool) (float64, error) {
	if len(pos) < x+1 {
		return 0, errors.New("invalid ddm data")
	}

	d, err := strconv.ParseInt(pos[0:x], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse degrees (%s): %w", pos[0:x], err)
	}

	dm, err := strconv.ParseFloat(pos[x:], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse decimal minutes (%s): %w", pos[x:], err)
	}

	dd := ddmtodd(float64(d), dm)

	if dir {
		dd *= -1.0
	}

	return dd, nil
}

func ddmtodd(d, dm float64) float64 {
	return d + (dm / 60)
}

func batteryConversion(in int) float64 {
	switch in {
	case 1:
		return 10
	case 2:
		return 20
	case 3:
		return 40
	case 4:
		return 60
	case 5:
		return 80
	case 6:
		return 100
	}

	return 0
}
