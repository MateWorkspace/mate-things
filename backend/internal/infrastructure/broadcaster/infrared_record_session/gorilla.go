package infrastructurebroadcasterinfraredrecordsession

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"sync"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
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
	sessionId uuid.UUID,
) (err error) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return domainmodels.NewError("failed to upgrade websocket connection", domainmodels.ErrTypeFailure, err)
	}

	c := newClient(conn, sessionId)
	h.add(c)
	defer h.remove(c)

	go c.writePump()
	c.readPump(ctx)

	return nil
}

func (h *gorillaImpl) Send(
	ctx context.Context,
	event domainmodels.InfraredRecordSessionEvent,
) (err error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return domainmodels.NewError("failed to marshal infrared record session event", domainmodels.ErrTypeFailure, err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.sessions {
		if c.sessionId == event.SessionId {
			c.enqueue(payload)
		}
	}

	return nil
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

var _ domaincontractsbroadcaster.InfraredRecordSession = (*gorillaImpl)(nil)
