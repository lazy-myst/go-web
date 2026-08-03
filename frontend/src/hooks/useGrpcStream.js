import { useEffect, useRef, useState } from "react";
import { streamMessages } from "../lib/grpc";

// Subscribes to the server-streaming StreamMessages RPC for as long as
// `token` is set. Unlike the previous socket.io hook, each fetch() call here
// is independent, so React 18 StrictMode's mount->cleanup->mount in dev just
// aborts one fetch and starts a fresh one — no shared-connection gotcha.
export function useGrpcStream(token, onMessage) {
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState(null);
  const onMessageRef = useRef(onMessage);

  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  useEffect(() => {
    if (!token) {
      setConnected(false);
      return undefined;
    }

    const controller = new AbortController();
    let cancelled = false;

    streamMessages(
      token,
      (msg) => onMessageRef.current?.(msg),
      controller.signal,
      () => {
        if (!cancelled) {
          setConnected(true);
          setError(null);
        }
      }
    )
      .catch((err) => {
        if (!cancelled && err.name !== "AbortError") {
          setError(err.message);
        }
      })
      .finally(() => {
        if (!cancelled) setConnected(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [token]);

  return { connected, error };
}
