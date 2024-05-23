package status

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/freman/gps2mqtt/mqtt"
)

type Connections struct {
	mu          sync.RWMutex
	connections map[net.Conn]*Connection
}

type Connection struct {
	LastTimestamp time.Time
	LastPacket    mqtt.Identifier
}

func (c *Connections) Connected(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connections[conn] = &Connection{}
}

func (c *Connections) Packet(conn net.Conn, packet mqtt.Identifier) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connections[conn].LastPacket = packet
	c.connections[conn].LastTimestamp = time.Now()
}

func (c *Connections) Disconnected(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, conn)
}

func (c *Connections) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	r := make([]jsonBlob, 0, len(c.connections))
	for c, p := range c.connections {
		r = append(r, jsonBlob{
			LocalAddr:     c.LocalAddr(),
			RemoteAddr:    c.RemoteAddr(),
			LastTimestamp: p.LastTimestamp,
			LastPacket:    p.LastPacket,
		})
	}

	return json.MarshalIndent(r, "", "\t")
}

type jsonBlob struct {
	LocalAddr     net.Addr
	RemoteAddr    net.Addr
	LastTimestamp time.Time
	LastPacket    mqtt.Identifier
}
