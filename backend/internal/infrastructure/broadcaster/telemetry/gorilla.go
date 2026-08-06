package infrastructurebroadcastertelemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"sync"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type gorillaImpl struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*client
	upgrader websocket.Upgrader
}

func NewGorillaImpl(
	allowedOrigins []string,
) *gorillaImpl {
	return &gorillaImpl{
		sessions: make(map[uuid.UUID]*client),
		upgrader: websocket.Upgrader{
			CheckOrigin: originChecker(allowedOrigins),
		},
	}
}

func (h *gorillaImpl) Register(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	userId uuid.UUID,
	nodeDeviceId *string,
	metricName *string,
	initial *domainmodels.TelemetryRecord,
) (err error) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return domainmodels.NewError("failed to upgrade websocket connection", domainmodels.ErrTypeFailure, err)
	}

	c := newClient(conn, userId, nodeDeviceId, metricName, r.RemoteAddr)
	h.add(c)
	defer h.remove(c)

	if initial != nil {
		if payload, marshalErr := json.Marshal(*initial); marshalErr == nil {
			c.enqueue(payload)
		}
	}

	go c.writePump()
	c.readPump(ctx)

	return nil
}

func (h *gorillaImpl) Send(
	ctx context.Context,
	record domainmodels.TelemetryRecord,
) (err error) {
	select {
	case <-ctx.Done():
		return domainmodels.NewError(
			"failed to broadcast due to context timeout",
			domainmodels.ErrTypeTimeout,
			ctx.Err(),
		)
	default:
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return domainmodels.NewError("failed to marshal telemetry broadcast payload", domainmodels.ErrTypeFailure, err)
	}

	h.mu.RLock()
	targets := make([]*client, 0, len(h.sessions))
	for _, c := range h.sessions {
		if c.matches(record.NodeDeviceId, record.MetricName) {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		c.enqueue(payload)
	}

	return nil
}

func (h *gorillaImpl) SessionList(ctx context.Context) (sessions []domainmodels.BroadcastSessionTelemetry, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions = make([]domainmodels.BroadcastSessionTelemetry, 0, len(h.sessions))
	for _, c := range h.sessions {
		sessions = append(sessions, domainmodels.BroadcastSessionTelemetry{
			Id:           c.id,
			UserId:       c.userId,
			NodeDeviceId: c.nodeDeviceId,
			MetricName:   c.metricName,
			RemoteAddr:   c.remoteAddr,
			ConnectedAt:  c.connectedAt,
		})
	}

	return sessions, nil
}

func (h *gorillaImpl) add(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[c.id] = c
}

func (h *gorillaImpl) remove(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, c.id)
}

func originChecker(allowedOrigins []string) func(r *http.Request) bool {
	allowAll := slices.Contains(allowedOrigins, "*")

	return func(r *http.Request) bool {
		if allowAll {
			return true
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		parsed, err := url.Parse(origin)
		if err != nil {
			return false
		}

		for _, allowed := range allowedOrigins {
			if allowed == parsed.Host || allowed == origin {
				return true
			}
		}

		return false
	}
}
