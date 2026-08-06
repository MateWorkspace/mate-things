package infrastructurebroadcastertelemetry

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	sendBufferSize = 16
	pingInterval   = 30 * time.Second
	pongWait       = 60 * time.Second
)

type client struct {
	id           uuid.UUID
	userId       uuid.UUID
	nodeDeviceId *string
	metricName   *string
	remoteAddr   string
	connectedAt  time.Time

	conn *websocket.Conn

	mu     sync.Mutex
	send   chan []byte
	closed bool
}

func newClient(conn *websocket.Conn, userId uuid.UUID, nodeDeviceId, metricName *string, remoteAddr string) *client {
	return &client{
		id:           uuid.New(),
		userId:       userId,
		nodeDeviceId: nodeDeviceId,
		metricName:   metricName,
		remoteAddr:   remoteAddr,
		connectedAt:  time.Now().UTC(),
		conn:         conn,
		send:         make(chan []byte, sendBufferSize),
	}
}

func (c *client) matches(nodeDeviceId, metricName string) bool {
	if c.nodeDeviceId != nil && *c.nodeDeviceId != nodeDeviceId {
		return false
	}
	if c.metricName != nil && *c.metricName != metricName {
		return false
	}
	return true
}

func (c *client) enqueue(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	select {
	case c.send <- message:
	default:
		c.closeLocked()
	}
}

func (c *client) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
}

func (c *client) closeLocked() {
	if c.closed {
		return
	}
	c.closed = true
	close(c.send)
	_ = c.conn.Close()
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.close()
				return
			}

		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.close()
				return
			}
		}
	}
}

func (c *client) readPump(ctx context.Context) {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := c.conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	c.close()
}
