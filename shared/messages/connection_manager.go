package messages

import (
	"errors"
	"net/http"
	"ride-sharing/shared/contracts"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type connWrapper struct {
	conn  *websocket.Conn
	mutex sync.RWMutex
}

type ConnectionManager struct {
	connections map[string]*connWrapper
	mutex       sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (cm *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*connWrapper),
	}
}

func (cm *ConnectionManager) Get(id string) (*websocket.Conn, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	wrapper, ok := cm.connections[id]
	if !ok {
		return nil, false
	}

	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	return wrapper.conn, true
}

func (cm *ConnectionManager) SendMessage(id string, message contracts.WSMessage) error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	wrapper, ok := cm.connections[id]
	if !ok {
		return ErrConnectionNotFound
	}

	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	return wrapper.conn.WriteJSON(message)
}

func (cm *ConnectionManager) Add(id string, conn *websocket.Conn) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.connections[id] = &connWrapper{
		conn: conn,
	}
}

func (cm *ConnectionManager) Remove(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.connections, id)
}
