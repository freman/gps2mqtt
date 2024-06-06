package huabao

import (
	"bufio"
	"bytes"
	"io"
)

type reader struct {
	*bufio.Reader
}

func (r reader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i], err = r.Reader.ReadByte()
		if err != nil {
			return i, err
		}

		if p[i] == 0x7d {
			tmp, err := r.Reader.ReadByte()
			if err != nil {
				return i, err
			}

			if tmp == 0x01 {
				p[i] = 0x7d
			} else if tmp == 0x02 {
				p[i] = 0x7e
			}
		}
	}

	return len(p), nil
}

type writer struct {
	io.Writer
}

func (w writer) Write(p []byte) (n int, err error) {
	var ob bytes.Buffer

	lastIndex := len(p) - 1

	for i := range p {
		if i == 0 || i == lastIndex {
			ob.WriteByte(p[i])
			continue
		}

		if p[i] == 0x7d {
			ob.WriteByte(0x7d)
			ob.WriteByte(0x01)
			continue
		}

		if p[i] == 0x7e {
			ob.WriteByte(0x7d)
			ob.WriteByte(0x02)
			continue
		}

		ob.WriteByte(p[i])
	}

	return w.Writer.Write(ob.Bytes())
}
