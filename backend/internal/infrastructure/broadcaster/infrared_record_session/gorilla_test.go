package infrastructurebroadcasterinfraredrecordsession

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func startListener(t *testing.T, h *gorillaImpl, sessionId uuid.UUID) (conn *websocket.Conn, stop func()) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		_ = h.Register(ctx, w, r, sessionId)
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("failed to dial: %v", err)
	}

	return conn, func() {
		cancel()
		conn.Close()
		<-done
		srv.Close()
	}
}

func TestRegisterRejectsSecondListenerForSameSession(t *testing.T) {
	h := NewGorillaImpl([]string{"*"})
	sessionId := uuid.New()

	_, stop := startListener(t, h, sessionId)
	defer stop()

	err := h.Register(context.Background(), nil, nil, sessionId)
	if err == nil {
		t.Fatal("expected an error registering a second listener for the same session")
	}
	if !errors.Is(err, domainmodels.ErrTypeBroadcastListenerLimitReached) {
		t.Fatalf("expected ErrTypeBroadcastListenerLimitReached, got: %v", err)
	}
}

func TestRegisterAllowsListenerAfterPreviousOneDisconnects(t *testing.T) {
	h := NewGorillaImpl([]string{"*"})
	sessionId := uuid.New()

	_, stop := startListener(t, h, sessionId)
	stop()

	conn2, stop2 := startListener(t, h, sessionId)
	defer stop2()

	if conn2 == nil {
		t.Fatal("expected the second listener to register successfully after the first disconnected")
	}
}

func TestRegisterAllowsListenersForDifferentSessions(t *testing.T) {
	h := NewGorillaImpl([]string{"*"})

	_, stop1 := startListener(t, h, uuid.New())
	defer stop1()

	_, stop2 := startListener(t, h, uuid.New())
	defer stop2()
}

func TestSendOnlyReachesTheRegisteredListener(t *testing.T) {
	h := NewGorillaImpl([]string{"*"})
	sessionId := uuid.New()

	conn, stop := startListener(t, h, sessionId)
	defer stop()

	event := domainmodels.InfraredRecordSessionEvent{SessionId: sessionId}
	if err := h.Send(context.Background(), event); err != nil {
		t.Fatalf("unexpected error sending event: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("expected to receive the broadcast event, got error: %v", err)
	}
}
