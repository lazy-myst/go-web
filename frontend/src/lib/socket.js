import io from "socket.io-client";

const SOCKET_URL = import.meta.env.VITE_SOCKET_URL || "http://localhost:3001";

let socket = null;

export function connectSocket(token) {
  if (!socket) {
    socket = io(SOCKET_URL, {
      query: { token },
      transports: ["websocket"],
      autoConnect: false,
    });
  } else {
    socket.io.opts.query = { token };
  }

  if (!socket.connected) {
    socket.connect();
  }

  return socket;
}

export function getSocket() {
  return socket;
}

export function disconnectSocket() {
  socket?.disconnect();
}
