package h02

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
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

	switch char {
	case '*':
		np, err := p.readAsciiPacket()
		if err != nil {
			return nil, fmt.Errorf("ascii error: %w", err)
		}
		return np, nil
	case '$':
		np, err := p.readBinaryPacket()
		if err != nil {
			return nil, fmt.Errorf("binary error: %w", err)
		}
		return np, nil
	}

	return nil, errors.New("bad packet")
}

// HQ,2204000054,V1,112752,A,2703.9008,S,15255.7160,E,000.00,027,220423,FFFFFBFF,mcc,mnc,ff64,0,6#
// 0, 1         ,2 ,3     ,4,5        ,6,7         ,8,9     ,10 ,11    ,12      ,13 ,14 ,15, 16,17

func (p *Parser) readAsciiPacket() (packet interface{}, err error) {
	var str string

	if str, err = p.readString('#'); err != nil {
		return packet, err
	}

	data := strings.Split(str, ",")
	switch data[0] + ":" + data[2] {
	case "HQ:V1": // Location
		np := PacketHQV1{
			&Packet{
				DeviceID: data[1],
				Position: strings.EqualFold(data[4], "A"),
			},
		}

		ts := data[11] + data[3]
		if np.Timestamp, err = time.Parse("020106150405", ts); err != nil {
			return nil, fmt.Errorf("%w (%s != 020106150405)", err, ts)
		}

		if np.Latitude, err = strLatitude(data[5], strings.EqualFold(data[6], "S")); err != nil {
			return nil, err
		}

		if np.Longitude, err = strLongitude(data[7], strings.EqualFold(data[8], "W")); err != nil {
			return nil, err
		}

		if np.Speed, err = strconv.ParseFloat(data[9], 64); err != nil {
			return nil, err
		}

		np.Speed *= 1.852 // Convert from knots to km/hr

		if data[10] != "" { // Null is 0 apparently
			if np.Heading, err = strconv.ParseFloat(strings.TrimLeft(data[10], "0"), 64); err != nil {
				return nil, err
			}
		}

		if data[17] != "" {
			tmpi, err := strconv.Atoi(data[17])
			if err != nil {
				return nil, err
			}

			np.Battery = batteryConversion(tmpi)
		}

		return np, nil
	case "HQ:V19": // Sim data
		return nil, nil
	}

	return nil, errors.New("bad packet")
}

func (p *Parser) readBinaryPacket() (packet interface{}, err error) {
	data := make([]byte, 50)

	c, err := p.reader.Read(data)
	if err != nil {
		return nil, err
	}

	if c < 50 {
		return nil, errors.New("not enough data")
	}

	np := PacketB{
		&Packet{
			DeviceID: hex.EncodeToString(data[0:5]),
			Battery:  batteryConversion(int(data[15])),
			Position: data[20]&2 == 2,
		},
	}

	ts := hex.EncodeToString(data[5:11])
	if np.Timestamp, err = time.Parse("150405060102", ts); err != nil {
		return nil, fmt.Errorf("%w (%s != 150405060102)", err, ts)
	}

	west := data[20]&8 == 8
	south := data[20]&4 == 4
	hexLongitude := hex.EncodeToString(data[16:21])

	if np.Longitude, err = strLongitude(hexLongitude[0:5]+"."+hexLongitude[5:9], south); err != nil {
		return nil, err
	}

	if np.Latitude, err = strLatitude(hex.EncodeToString(data[11:13])+"."+hex.EncodeToString(data[13:15]), west); err != nil {
		return nil, err
	}

	hexSpeedDir := hex.EncodeToString(data[21:24])
	if np.Speed, err = strconv.ParseFloat(hexSpeedDir[0:3], 64); err != nil {
		return nil, err
	}

	np.Speed *= 1.852 // Convert from knots to km/hr

	if np.Heading, err = strconv.ParseFloat(strings.TrimLeft(hexSpeedDir[3:6], "0"), 64); err != nil {
		return nil, err
	}

	return np, nil
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
		return 0, errors.New("invalid data")
	}

	d, err := strconv.ParseInt(pos[0:x], 10, 64)
	if err != nil {
		return 0, err
	}

	dm, err := strconv.ParseFloat(pos[x:], 64)
	if err != nil {
		return 0, err
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
