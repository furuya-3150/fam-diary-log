package ws

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type connection struct {
	ws        *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
}

func newConnection(ws *websocket.Conn) *connection {
	return &connection{ws: ws, send: make(chan []byte, 1), done: make(chan struct{})}
}

func (c *connection) reader() {
	slog.Debug("connection reader: started")
	defer func() {
		slog.Debug("connection reader: stopping")
		c.close()
	}()
	for {
		if _, _, err := c.ws.NextReader(); err != nil {
			slog.Debug("connection reader: NextReader error", "error", err)
			return
		}
		// ignore received messages
	}
}

func (c *connection) writer() {
	slog.Debug("connection writer: started")
	for b := range c.send {
		c.mu.Lock()
		slog.Debug("connection writer: sending message", "message", string(b))
		_ = c.ws.WriteMessage(websocket.TextMessage, b)
		c.mu.Unlock()
	}
	slog.Debug("connection writer: channel closed, stopping")
	c.close()
}

func (c *connection) close() {
	c.closeOnce.Do(func() {
		slog.Debug("connection close: closing connection")
		c.mu.Lock()
		if c.ws != nil {
			_ = c.ws.Close()
			c.ws = nil
		}
		c.mu.Unlock()
		// signal done
		select {
		case <-c.done:
			slog.Debug("connection close: done already closed")
		default:
			slog.Debug("connection close: closing done channel")
			close(c.done)
		}
	})
}
