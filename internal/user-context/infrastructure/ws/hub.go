package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(ctx context.Context, userID uuid.UUID, payload interface{}) error
	CloseUserConnections(userID uuid.UUID)
}

type subscription struct {
	userID uuid.UUID
	conn   *connection
}

type message struct {
	userID  uuid.UUID
	payload []byte
}

// Hub keeps track of connections per user and broadcasts messages to them.
type Hub struct {
	register   chan subscription
	unregister chan subscription
	broadcast  chan message
	closeUser  chan uuid.UUID
	clients    map[uuid.UUID]map[*connection]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan subscription),
		unregister: make(chan subscription),
		broadcast:  make(chan message),
		closeUser:  make(chan uuid.UUID),
		clients:    make(map[uuid.UUID]map[*connection]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case s := <-h.register:
			slog.Debug("Hub Run: registering connection", "user_id", s.userID)
			conns := h.clients[s.userID]
			if conns == nil {
				conns = make(map[*connection]struct{})
				h.clients[s.userID] = conns
			}
			h.clients[s.userID][s.conn] = struct{}{}
		case s := <-h.unregister:
			slog.Debug("Hub Run: unregistering connection", "user_id", s.userID)
			if conns, ok := h.clients[s.userID]; ok {
				delete(conns, s.conn)
				s.conn.close()
				if len(conns) == 0 {
					delete(h.clients, s.userID)
				}
			}
		case m := <-h.broadcast:
			slog.Debug("Hub Run: broadcasting message", "user_id", m.userID, "payload", string(m.payload))
			if conns, ok := h.clients[m.userID]; ok {
				for c := range conns {
					select {
					case c.send <- m.payload:
					default:
						c.close()
						delete(conns, c)
					}
				}
			}
		case uid := <-h.closeUser:
			slog.Debug("Hub Run: closing all connections for user", "user_id", uid)
			if conns, ok := h.clients[uid]; ok {
				for c := range conns {
					c.close()
					delete(conns, c)
				}
				delete(h.clients, uid)
			}
		}
	}
}

func (h *Hub) Publish(ctx context.Context, userID uuid.UUID, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	h.broadcast <- message{userID: userID, payload: b}
	return nil
}

// CloseUserConnections closes all WebSocket connections for the specified user.
func (h *Hub) CloseUserConnections(userID uuid.UUID) {
	h.closeUser <- userID
}
