package huabao

import (
	"bufio"
	"errors"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/freman/gps2mqtt/mqtt"
	"github.com/freman/gps2mqtt/protocol"
	"github.com/freman/gps2mqtt/status"
	"github.com/rs/zerolog"
)

type Listener struct {
	log       zerolog.Logger
	whitelist func(name string) bool

	connections *status.Connections

	Listen       string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func (l *Listener) Run(chMsg chan mqtt.Identifier) error {
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
		l.connections.Connected(c)

		go l.HandleConnection(c, chMsg, logger)
	}
}

func (l *Listener) HandleConnection(c net.Conn, chMsg chan mqtt.Identifier, log zerolog.Logger) {
	defer func() {
		log.Info().Msg("Client disconnected.")
		l.connections.Disconnected(c)

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

		if !l.CheckWhitelist(packet) {
			log.Warn().Str("device", packet.Device()).Msg("Rejecting unknown device.")
			c.Close()

			return
		}

		l.connections.Packet(c, packet)

		if packet.WantResponse() {
			if err := c.SetWriteDeadline(time.Now().Add(l.WriteTimeout)); err != nil {
				log.Error().Err(err).Msg("Failed to set a write deadline.")
				return
			}

			if err := packet.Respond(c); err != nil {
				log.Error().Err(err).Msg("Failed to finish handshake.")
			}

			if err := c.SetWriteDeadline(time.Time{}); err != nil {
				log.Error().Err(err).Msg("Failed to clear a write deadline.")
				return
			}
		}

		chMsg <- packet
	}
}

func (l *Listener) CheckWhitelist(p *Packet) bool {
	return l.whitelist(p.Device())
}

func (l *Listener) Setup(config protocol.Configerer) error {
	l.Listen = ":5015"
	l.WriteTimeout = time.Second
	l.ReadTimeout = time.Minute

	if err := config.ProtocolConfiguration(Name, l); err != nil {
		return err
	}

	l.whitelist = config.Whitelist
	l.connections = status.NewConnections(Name)

	return nil
}

func init() {
	protocol.Register(Name, func(logger zerolog.Logger) protocol.Interface {
		return &Listener{
			log: logger,
		}
	})
}
