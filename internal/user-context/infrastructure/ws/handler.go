package ws

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// basic origin check: allow same host origin
		origin := r.Header.Get("Origin")
		slog.Debug("WS Upgrader CheckOrigin", "origin", origin)
		if origin == "" {
			return false
		}
		host := r.Host
		slog.Debug("WS Upgrader CheckOrigin", "host", host)
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
	val := c.Request().Context().Value(auth.ContextKeyUserID)
	ctxUserID, ok := val.(uuid.UUID)
	if !ok {
		slog.Error("WSHandler Handle: invalid user_id context", "user_id", val)
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	wsConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		slog.Error("WSHandler Handle: failed to upgrade websocket connection", "error", err)
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "failed to upgrade websocket connection"})
	}
	conn := newConnection(wsConn)
	sub := subscription{userID: ctxUserID, conn: conn}
	h.hub.register <- sub

	slog.Info("WebSocket handshake successful", "user_id", ctxUserID, "status", 101)

	// start reader/writer
	go conn.writer()
	go conn.reader()

	// block until connection is closed
	<-conn.done
	h.hub.unregister <- sub

	return nil
}
