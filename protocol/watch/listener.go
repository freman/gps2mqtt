package watch

import (
	"bufio"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/freman/gps2mqtt/mqtt"
	"github.com/freman/gps2mqtt/protocol"
	"github.com/rs/zerolog"
)

type Listener struct {
	log       zerolog.Logger
	whitelist func(name string) bool

	Listen       string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func (l *Listener) Run(chMsg chan mqtt.Message) error {
	l.log.Info().Str("cfg", l.Listen).Msg("Starting listener.")

	nl, err := net.Listen("tcp", l.Listen)
	if err != nil {
		panic(err)
	}

	for {
		c, err := nl.Accept()

		if err != nil {
			l.log.Fatal().Err(err).Msg("Unable to start listener.")
		}

		logger := l.log.With().IPAddr("remote", c.RemoteAddr().(*net.TCPAddr).IP).Logger()
		logger.Info().Msg("Client connected.")

		go l.HandleConnection(c, chMsg, logger)
	}
}

func (l *Listener) HandleConnection(c net.Conn, chMsg chan mqtt.Message, log zerolog.Logger) {
	defer func() {
		if err := c.Close(); err != nil {
			log.Error().Err(err).Msg("Error while closing client connection.")
		}
	}()

	p := Parser{
		reader: bufio.NewReader(c),
	}

	for {
		if err := c.SetReadDeadline(time.Now().Add(l.ReadTimeout)); err != nil {
			log.Error().Err(err).Msg("Failed to set a read deadline.")
			return
		}

		packet, err := p.ReadPacket()
		if err != nil {
			if err == io.EOF || errors.Is(err, syscall.ECONNRESET) {
				return
			}

			log.Error().Err(err).Msg("Failed to read packet.")

			return
		}

		if err := c.SetReadDeadline(time.Time{}); err != nil {
			log.Error().Err(err).Msg("Failed to clear a read deadline.")
			return
		}

		switch p := packet.(type) {
		case *PacketLK:
			if !l.CheckWhitelist(p) {
				c.Close()
				return
			}

			if err := c.SetWriteDeadline(time.Now().Add(l.WriteTimeout)); err != nil {
				log.Error().Err(err).Msg("Failed to set a write deadline.")
				return
			}

			if err := p.Respond(c); err != nil {
				log.Error().Err(err).Msg("Failed to finish handshake.")
			}

			if err := c.SetWriteDeadline(time.Time{}); err != nil {
				log.Error().Err(err).Msg("Failed to clear a write deadline.")
				return
			}

			chMsg <- mqtt.Message{
				Data: p,
				Type: mqtt.TypeHello,
			}
		case *PacketUD:
			chMsg <- mqtt.Message{
				Data: p,
				Type: mqtt.TypeUpdate,
			}
		}
	}
}

func (l *Listener) CheckWhitelist(p *PacketLK) bool {
	return l.whitelist(p.Device())
}

func (l *Listener) Setup(config protocol.Configerer) error {
	l.Listen = ":5093"
	l.WriteTimeout = time.Second
	l.ReadTimeout = time.Minute

	if err := config.ProtocolConfiguration("watch", l); err != nil {
		return err
	}

	l.whitelist = config.Whitelist

	return nil
}

func init() {
	protocol.Register("watch", func(logger zerolog.Logger) protocol.Interface {
		return &Listener{
			log: logger,
		}
	})
}
