package status

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
)

type status struct {
	mu sync.RWMutex

	connections map[string]*Connections
}

var s = status{
	connections: map[string]*Connections{},
}

func NewConnections(proto string) *Connections {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := &Connections{
		connections: make(map[net.Conn]*Connection),
	}

	s.connections[proto] = c

	return c
}

func (s *status) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(s.connections)
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.connections)
}
