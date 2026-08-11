"use client";

import { useEffect, useRef, useState } from "react";

const MAX_RECONNECT_ATTEMPTS = 5;
const BASE_RECONNECT_DELAY_MS = 1000;

// Only one WebSocket listener is allowed per session on the backend
// (gorilla.go's reserve() rejects a second concurrent registration) - the
// exponential backoff here exists so a stuck/slow-to-release server-side
// listener isn't hammered with reconnect attempts. After MAX_RECONNECT_
// ATTEMPTS the caller is expected to fall back to interval polling.
export function useInfraredRecordSessionBroadcast(
  sessionId: string,
  onEvent: () => void,
  getToken: () => Promise<string | null>,
): { connected: boolean; exhausted: boolean } {
  const [connected, setConnected] = useState(false);
  const [exhausted, setExhausted] = useState(false);
  const onEventRef = useRef(onEvent);
  const getTokenRef = useRef(getToken);
  const attemptRef = useRef(0);

  useEffect(() => {
    onEventRef.current = onEvent;
    getTokenRef.current = getToken;
  }, [onEvent, getToken]);

  useEffect(() => {
    let cancelled = false;
    let socket: WebSocket | null = null;
    let retryTimeout: number | undefined;
    attemptRef.current = 0;

    async function connect() {
      let token: string | null;
      try {
        token = await getTokenRef.current();
      } catch {
        token = null;
      }
      if (cancelled) {
        return;
      }
      if (!token) {
        setExhausted(true);
        return;
      }

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${protocol}//${window.location.host}/api/v1/infrared/record-sessions/${sessionId}/broadcast?token=${encodeURIComponent(token)}`;
      socket = new WebSocket(url);

      socket.onopen = () => {
        if (cancelled) return;
        attemptRef.current = 0;
        setConnected(true);
        setExhausted(false);
      };

      socket.onmessage = () => {
        if (!cancelled) {
          onEventRef.current();
        }
      };

      socket.onclose = () => {
        if (cancelled) return;
        setConnected(false);
        if (attemptRef.current >= MAX_RECONNECT_ATTEMPTS) {
          setExhausted(true);
          return;
        }
        const delay = BASE_RECONNECT_DELAY_MS * 2 ** attemptRef.current;
        attemptRef.current += 1;
        retryTimeout = window.setTimeout(connect, delay);
      };

      socket.onerror = () => {
        socket?.close();
      };
    }

    connect();

    return () => {
      cancelled = true;
      if (retryTimeout) {
        window.clearTimeout(retryTimeout);
      }
      socket?.close();
    };
  }, [sessionId]);

  return { connected, exhausted };
}
