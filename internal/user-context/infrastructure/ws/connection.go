package ws

import (
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
	defer func() { c.close() }()
	for {
		if _, _, err := c.ws.NextReader(); err != nil {
			return
		}
		// ignore received messages
	}
}

func (c *connection) writer() {
	for b := range c.send {
		c.mu.Lock()
		_ = c.ws.WriteMessage(websocket.TextMessage, b)
		c.mu.Unlock()
	}
	c.close()
}

func (c *connection) close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		if c.ws != nil {
			_ = c.ws.Close()
			c.ws = nil
		}
		c.mu.Unlock()
		// signal done
		select {
		case <-c.done:
			// already closed
		default:
			close(c.done)
		}
	})
}
