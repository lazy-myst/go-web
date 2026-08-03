import { useEffect, useRef, useState } from "react";
import { connectSocket, disconnectSocket } from "../lib/socket";

export function useSocket(token) {
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState(null);
  const socketRef = useRef(null);

  useEffect(() => {
    if (!token) {
      disconnectSocket();
      socketRef.current = null;
      return undefined;
    }

    const socket = connectSocket(token);
    socketRef.current = socket;

    const handleConnect = () => {
      setConnected(true);
      setError(null);
    };
    const handleDisconnect = () => setConnected(false);
    const handleConnectError = (err) => setError(err.message);

    socket.on("connect", handleConnect);
    socket.on("disconnect", handleDisconnect);
    socket.on("connect_error", handleConnectError);

    // Deliberately not disconnecting here: React 18 StrictMode runs this
    // effect's cleanup and re-run back-to-back in dev, and disconnecting an
    // in-flight connection aborts the handshake ("closed before the
    // connection is established"). The socket is a singleton (see socket.js)
    // that persists across remounts and is torn down when the token clears.
    return () => {
      socket.off("connect", handleConnect);
      socket.off("disconnect", handleDisconnect);
      socket.off("connect_error", handleConnectError);
    };
  }, [token]);

  return { socket: socketRef.current, connected, error };
}
