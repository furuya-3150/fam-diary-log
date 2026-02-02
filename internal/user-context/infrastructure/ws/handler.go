package ws

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// basic origin check: allow same host origin
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		host := r.Host
		if strings.HasPrefix(origin, "https://"+host) || strings.HasPrefix(origin, "http://"+host) {
			return true
		}
		return false
	},
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(h *Hub) *WSHandler {
	return &WSHandler{hub: h}
}

func (h *WSHandler) Handle(c echo.Context) error {
	// If user_id present in request context, ensure it matches token subject
	val := c.Request().Context().Value("user_id")
	ctxUserID, ok := val.(uuid.UUID);
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}

	wsConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	conn := newConnection(wsConn)
	sub := subscription{userID: ctxUserID, conn: conn}
	h.hub.register <- sub

	// start reader/writer
	go conn.writer()
	go conn.reader()

	// ensure unregister when connection is closed
	go func() {
		<-conn.done
		h.hub.unregister <- sub
	}()

	return nil
}
