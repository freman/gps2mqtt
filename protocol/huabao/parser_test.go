package huabao

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	fromGPS := []byte{0x7e,
		0x01, 0x00,
		0x00, 0x2d,
		0x01, 0x91, 0x75, 0x69, 0x02, 0x32,
		0x00, 0x07,
		0x00, 0x22, 0x04, 0x4e, 0x37, 0x30, 0x34, 0x34, 0x34, 0x4d, 0x4c, 0x35, 0x30, 0x30, 0x5f, 0x45, 0x44, 0x5f, 0x47, 0x54, 0x32, 0x35, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x35, 0x36, 0x39, 0x30, 0x32, 0x33, 0x32, 0x02, 0xd4, 0xc1, 0x42, 0x35, 0x37, 0x31, 0x39, 0x31,
		0xf2, // checksum
		0x7e,
	}

	toGPS := []byte{0x7e,
		0x81, 0x00, // message id
		0x00, 0x0f, // properties
		0x01, 0x91, 0x75, 0x69, 0x02, 0x32, // terminal
		0x00, 0x00, // sequence
		0x00, 0x07, // response sequence
		0x00, // status byte
		0x30, 0x31, 0x39, 0x31, 0x37, 0x35, 0x36, 0x39, 0x30, 0x32, 0x33, 0x32,
		0x32, // checksum
		0x7e,
	}

	p := &Parser{reader: bufio.NewReader(bytes.NewReader(fromGPS))}
	packet, err := p.ReadPacket()
	assert.NoError(t, err)

	var r bytes.Buffer
	assert.NoError(t, packet.Respond(&r))
	assert.Equal(t, toGPS, r.Bytes())

	assert.Equal(t, "019175690232", packet.DeviceID)
	assert.Equal(t, time.Time{}, packet.Timestamp)
	assert.Equal(t, 0.0, packet.Latitude)
	assert.Equal(t, 0.0, packet.Longitude)
	assert.Equal(t, 0.0, packet.Altitude)
	assert.Equal(t, 0.0, packet.Heading)
	assert.Equal(t, 0.0, packet.Speed)
	assert.Equal(t, false, packet.Position)
	assert.Equal(t, int64(0), packet.Satellites)
	assert.Equal(t, 0.0, packet.RSSI)
	assert.Equal(t, 0.0, packet.Battery)
	assert.Equal(t, "70444", packet.ManufacturerID)
	assert.Equal(t, "ML500_ED_GT25H", packet.TerminalModel)
	assert.Equal(t, "5690232", packet.TerminalID)
}

func TestAuthenticate(t *testing.T) {
	fromGPS := []byte{0x7e, 0x01, 0x02, 0x00, 0x0c, 0x01, 0x91, 0x75, 0x69, 0x02, 0x32, 0x00, 0x09, 0x30, 0x31, 0x39, 0x31, 0x37, 0x35, 0x36, 0x39, 0x30, 0x32, 0x33, 0x32, 0xbd, 0x7e}
	toGPS := []byte{0x7e,
		0x80, 0x01, //message id
		0x00, 0x05, //properties
		0x01, 0x91, 0x75, 0x69, 0x02, 0x32, //terminal
		0x00, 0x00, // sequence --
		0x00, 0x09, // response sequence
		0x01, 0x02, // original message id
		0x00, // code
		0x32, // crc
		0x7e,
	}

	p := &Parser{reader: bufio.NewReader(bytes.NewReader(fromGPS))}
	packet, err := p.ReadPacket()
	assert.NoError(t, err)

	var r bytes.Buffer
	assert.NoError(t, packet.Respond(&r))
	assert.Equal(t, toGPS, r.Bytes())

	assert.Equal(t, "019175690232", packet.DeviceID)
	assert.Equal(t, time.Time{}, packet.Timestamp)
	assert.Equal(t, 0.0, packet.Latitude)
	assert.Equal(t, 0.0, packet.Longitude)
	assert.Equal(t, 0.0, packet.Altitude)
	assert.Equal(t, 0.0, packet.Heading)
	assert.Equal(t, 0.0, packet.Speed)
	assert.Equal(t, false, packet.Position)
	assert.Equal(t, int64(0), packet.Satellites)
	assert.Equal(t, 0.0, packet.RSSI)
	assert.Equal(t, 0.0, packet.Battery)
	assert.Equal(t, "", packet.ManufacturerID)
	assert.Equal(t, "", packet.TerminalModel)
	assert.Equal(t, "", packet.TerminalID)
}

func TestLocationReport(t *testing.T) {
	fromGPS := []byte{0x7e, 0x02, 0x00, 0x00, 0x56, 0x01, 0x91, 0x75, 0x69, 0x02, 0x32, 0x00, 0xb9, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x02, 0x9c, 0xfa, 0xee, 0x08, 0x1d, 0x81, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x24, 0x06, 0x04, 0x01, 0x28, 0x23, 0x01, 0x04, 0x00, 0x00, 0x00, 0x65, 0x30, 0x01, 0x0f, 0x31, 0x01, 0x00, 0x51, 0x02, 0x00, 0x00, 0x57, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x9f, 0x17, 0x35, 0x30, 0x35, 0x2c, 0x30, 0x31, 0x2c, 0x37, 0x30, 0x30, 0x64, 0x2c, 0x30, 0x38, 0x63, 0x62, 0x34, 0x61, 0x32, 0x39, 0x2c, 0x31, 0x35, 0xe1, 0x01, 0x64, 0xe2, 0x02, 0x00, 0x00, 0xf4, 0x7e}
	toGPS := []byte{0x7e, 0x80, 0x01, 0x00, 0x05, 0x01, 0x91, 0x75, 0x69, 0x02, 0x32, 0x00, 0x00, 0x00, 0xb9, 0x02, 0x00, 0x00, 0x83, 0x7e}

	p := &Parser{reader: bufio.NewReader(bytes.NewReader(fromGPS))}
	packet, err := p.ReadPacket()
	assert.NoError(t, err)

	var r bytes.Buffer
	assert.NoError(t, packet.Respond(&r))
	assert.Equal(t, toGPS, r.Bytes())

	assert.Equal(t, "019175690232", packet.DeviceID)
	assert.Equal(t, time.Date(2024, 06, 04, 01, 28, 23, 0, time.FixedZone("GMT+8", 8*60*60)), packet.Timestamp)
	assert.Equal(t, -43.842286, packet.Latitude)
	assert.Equal(t, 136.15134, packet.Longitude)
	assert.Equal(t, 0.0, packet.Altitude)
	assert.Equal(t, 0.0, packet.Heading)
	assert.Equal(t, 0.0, packet.Speed)
	assert.Equal(t, false, packet.Position)
	assert.Equal(t, int64(0), packet.Satellites)
	assert.Equal(t, 15.0, packet.RSSI)
	assert.Equal(t, 100.0, packet.Battery)
}

func Test7DDecode(t *testing.T) {
	fromGPS, _ := hex.DecodeString("7e02000056019175690232007d010000000000000000000000000000000000000000000024060316074101040000000030011931010051020000570800000000000000009f173530352c30312c373030642c30386338353030322c3235e10164e2020000287e")

	p := &Parser{reader: bufio.NewReader(bytes.NewReader(fromGPS))}
	packet, err := p.ReadPacket()
	assert.NoError(t, err)

	assert.Equal(t, "019175690232", packet.DeviceID)
	assert.Equal(t, time.Date(2024, 06, 03, 16, 07, 41, 0, time.FixedZone("GMT+8", 8*60*60)), packet.Timestamp)
	assert.Equal(t, 0.0, packet.Latitude)
	assert.Equal(t, 0.0, packet.Longitude)
	assert.Equal(t, 0.0, packet.Altitude)
	assert.Equal(t, 0.0, packet.Heading)
	assert.Equal(t, 0.0, packet.Speed)
	assert.Equal(t, false, packet.Position)
	assert.Equal(t, int64(0), packet.Satellites)
	assert.Equal(t, 25.0, packet.RSSI)
	assert.Equal(t, 100.0, packet.Battery)
	assert.Equal(t, "", packet.ManufacturerID)
	assert.Equal(t, "", packet.TerminalModel)
	assert.Equal(t, "", packet.TerminalID)
}

func TestFromTraccarLog(t *testing.T) {
	f, err := os.Open("test_data/tracker-server.log")
	if err != nil {
		t.Skip("Unable to open tracker-server.log")
	}

	defer f.Close()

	lineIndex := 0
	re := regexp.MustCompile(`INFO: \[\w+: huabao ([<>]) .+?\] ([0-9a-f]+)`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		lineIndex++

		matches := re.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 || matches[0][1] != "<" {
			continue
		}

		fromGPSString := matches[0][2]
		var toGPSString string

		for scanner.Scan() {
			line := scanner.Text()

			lineIndex++

			matches := re.FindAllStringSubmatch(line, -1)
			if len(matches) == 0 {
				continue
			}

			if matches[0][1] != ">" {
				if matches[0][2][0:2] != "7e" && !strings.HasSuffix(fromGPSString, "7e") {
					t.Logf("Packet split across multiple lines")
					fromGPSString += matches[0][2]

					continue
				}

				t.Logf("Didn't find reply packet: %d", lineIndex)
				fromGPSString = matches[0][2]

				continue
			}

			toGPSString = matches[0][2]
			break
		}

		fromGPS, err := hex.DecodeString(fromGPSString)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		toGPS, err := hex.DecodeString(toGPSString)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		p := &Parser{reader: bufio.NewReader(bytes.NewReader(fromGPS))}
		packet, err := p.ReadPacket()
		assert.NoError(t, err, lineIndex)

		var r bytes.Buffer
		assert.NoError(t, packet.Respond(&r))
		assert.Equal(t, toGPS, r.Bytes())

		if packet.Valid() {
			t.Logf("%v: Altitude: %0.2f, Latitude: %0.5f Longitude: %0.5f, Heading: %0.2f, Speed: %0.2f, Battery: %0.2f", packet.Timestamp, packet.Altitude, packet.Latitude, packet.Longitude, packet.Heading, packet.Speed, packet.Battery)
		}
	}
}
