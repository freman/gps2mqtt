package huabao

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Parser struct {
	reader *bufio.Reader

	terminalInfo struct {
		ManufacturerID [5]byte
		Model          [20]byte
		ID             [7]byte
	}
}

func (p *Parser) read(d []byte) (err error) {
	sz, err := reader{p.reader}.Read(d)
	if err != nil {
		return err
	}

	if len(d) != sz {
		return errors.New("unexpected number of bytes read")
	}

	return nil
}

func (p *Parser) ReadPacket() (packet *Packet, err error) {
	preamble, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if preamble != 0x7e {
		return nil, errors.New("bad preamble")
	}

	var head header
	if err := binary.Read(reader{p.reader}, binary.BigEndian, &head); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	if head.Properties.Encrypted() {
		return nil, errors.New("encryption not supported")
	}

	if head.Properties.SubPackage() {
		return nil, errors.New("subpackage not supported")
	}

	body := make([]byte, head.Properties.Length())
	p.read(body)

	checksum := make([]byte, 1)
	if err := p.read(checksum); err != nil {
		return nil, errors.New("unable to read checksum")
	}
	_ = checksum //todo check checksum

	tail, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if tail != 0x7e {
		return nil, errors.New("invalid tail")
	}

	bodyBuf := bytes.NewReader(body)

	packet = &Packet{
		header:   head,
		DeviceID: head.Terminal.String(),
	}

	switch head.MessageType {
	case protoRegister:
		bodyBuf.Seek(4, io.SeekCurrent)
		if err := binary.Read(bodyBuf, binary.BigEndian, &p.terminalInfo); err != nil {
			return nil, err
		}
	case protoLocationReport:
		var rep basicLocationInformation
		binary.Read(bodyBuf, binary.BigEndian, &rep)
		packet.importBasicInformation(rep)

		if err := p.parseAdditionalInformation(bodyBuf, packet); err != nil {
			return nil, err
		}
	}

	packet.ManufacturerID = strings.TrimRight(string(p.terminalInfo.ManufacturerID[:]), "\x00")
	packet.TerminalID = strings.TrimRight(string(p.terminalInfo.ID[:]), "\x00")
	packet.TerminalModel = strings.TrimRight(string(p.terminalInfo.Model[:]), "\x00")

	return packet, nil
}

func (p *Parser) parseAdditionalInformation(buf *bytes.Reader, packet *Packet) (err error) {
	for buf.Len() > 0 {
		addInfo, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("unable to read additinal information type: %w", err)
		}
		addLen, err := buf.ReadByte()
		if err != nil {
			return fmt.Errorf("unable to read additinal information len: %w", err)
		}

		switch addInfo {
		case 0x30:
			tmp, err := buf.ReadByte()
			if err != nil {
				return fmt.Errorf("unable to read RSSI from additional information: %w", err)
			}

			packet.RSSI = float64(tmp)
		case 0x31:
			tmp, err := buf.ReadByte()
			if err != nil {
				return fmt.Errorf("unable to read Satellites from additional information: %w", err)
			}

			packet.Satellites = int64(tmp)
		case 0xd4: // LT-160
			tmp, err := buf.ReadByte()
			if err != nil {
				return fmt.Errorf("unable to read battery percentage from additional information: %w", err)
			}

			packet.Battery = float64(tmp)
		case 0xe1: // ML100G
			tmp, err := buf.ReadByte()
			if err != nil {
				return fmt.Errorf("unable to read battery percentage from additional information: %w", err)
			}

			packet.Battery = float64(tmp)
		default:
			buf.Seek(int64(addLen), io.SeekCurrent) // skip
		}
	}

	return err
}
