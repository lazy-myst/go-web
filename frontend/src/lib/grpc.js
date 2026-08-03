// Minimal gRPC-Web client for the chat service, hand-written against the
// wire format directly (no generated stubs, no extra npm dependencies).
// Every field in our .proto messages is a string, so encode/decode only
// needs to handle the length-delimited (wire type 2) case.

const GRPC_URL = import.meta.env.VITE_GRPC_URL || "";

function encodeVarint(value) {
  const bytes = [];
  while (value > 0x7f) {
    bytes.push((value & 0x7f) | 0x80);
    value >>>= 7;
  }
  bytes.push(value & 0x7f);
  return bytes;
}

function decodeVarintAt(bytes, start) {
  let result = 0;
  let shift = 0;
  let i = start;
  for (;;) {
    const b = bytes[i];
    result |= (b & 0x7f) << shift;
    i++;
    if ((b & 0x80) === 0) break;
    shift += 7;
  }
  return [result >>> 0, i - start];
}

function encodeStringField(fieldNumber, value) {
  if (!value) return [];
  const tag = (fieldNumber << 3) | 2;
  const utf8 = Array.from(new TextEncoder().encode(value));
  return [tag, ...encodeVarint(utf8.length), ...utf8];
}

function encodeSendMessageRequest({ chatId, text }) {
  return new Uint8Array([
    ...encodeStringField(1, chatId),
    ...encodeStringField(2, text),
  ]);
}

const MESSAGE_FIELDS = {
  1: "id",
  2: "chatId",
  3: "senderId",
  4: "text",
  5: "createdAt",
};

function decodeMessage(bytes) {
  const result = {};
  let i = 0;
  while (i < bytes.length) {
    const [tag, tagLen] = decodeVarintAt(bytes, i);
    i += tagLen;
    const fieldNumber = tag >>> 3;
    const wireType = tag & 0x7;

    if (wireType === 2) {
      const [len, lenLen] = decodeVarintAt(bytes, i);
      i += lenLen;
      const slice = bytes.slice(i, i + len);
      i += len;
      const name = MESSAGE_FIELDS[fieldNumber];
      if (name) result[name] = new TextDecoder().decode(slice);
    } else if (wireType === 0) {
      const [, len] = decodeVarintAt(bytes, i);
      i += len;
    } else {
      break; // unsupported wire type in this message; stop rather than misparse
    }
  }
  return result;
}

function frameMessage(payloadBytes) {
  const framed = new Uint8Array(5 + payloadBytes.length);
  const len = payloadBytes.length;
  framed[0] = 0; // uncompressed
  framed[1] = (len >>> 24) & 0xff;
  framed[2] = (len >>> 16) & 0xff;
  framed[3] = (len >>> 8) & 0xff;
  framed[4] = len & 0xff;
  framed.set(payloadBytes, 5);
  return framed;
}

function extractFrames(buffer) {
  const frames = [];
  let offset = 0;
  while (buffer.length - offset >= 5) {
    const flag = buffer[offset];
    const len =
      ((buffer[offset + 1] << 24) |
        (buffer[offset + 2] << 16) |
        (buffer[offset + 3] << 8) |
        buffer[offset + 4]) >>>
      0;
    if (buffer.length - offset - 5 < len) break;
    const payload = buffer.slice(offset + 5, offset + 5 + len);
    frames.push({ isTrailer: (flag & 0x80) !== 0, payload });
    offset += 5 + len;
  }
  return { frames, rest: buffer.slice(offset) };
}

function concatBytes(a, b) {
  const out = new Uint8Array(a.length + b.length);
  out.set(a, 0);
  out.set(b, a.length);
  return out;
}

function grpcHeaders(token) {
  return {
    "Content-Type": "application/grpc-web+proto",
    "X-Grpc-Web": "1",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

function readTrailerStatus(payload) {
  const text = new TextDecoder().decode(payload);
  const statusMatch = text.match(/grpc-status:\s*(\d+)/i);
  const status = statusMatch ? parseInt(statusMatch[1], 10) : 0;
  if (status !== 0) {
    const msgMatch = text.match(/grpc-message:\s*(.*)/i);
    throw new Error(
      msgMatch ? decodeURIComponent(msgMatch[1].trim()) : `gRPC error ${status}`
    );
  }
}

async function unaryCall(method, requestBytes, token) {
  const res = await fetch(`${GRPC_URL}${method}`, {
    method: "POST",
    headers: grpcHeaders(token),
    body: frameMessage(requestBytes),
  });
  if (!res.ok) {
    throw new Error(`gRPC-Web request failed: HTTP ${res.status}`);
  }
  const buf = new Uint8Array(await res.arrayBuffer());
  const { frames } = extractFrames(buf);
  const messageFrame = frames.find((f) => !f.isTrailer);
  const trailerFrame = frames.find((f) => f.isTrailer);
  if (trailerFrame) readTrailerStatus(trailerFrame.payload);
  return messageFrame ? messageFrame.payload : null;
}

export async function sendMessage(token, chatId, text) {
  const respBytes = await unaryCall(
    "/chat.ChatService/SendMessage",
    encodeSendMessageRequest({ chatId, text }),
    token
  );
  return respBytes ? decodeMessage(respBytes) : null;
}

// Opens a server-streaming call and invokes onMessage for each Message the
// server pushes, until the stream ends or `signal` aborts it. Calls
// onOpen once the connection is actually established (HTTP headers back).
export async function streamMessages(token, onMessage, signal, onOpen) {
  const res = await fetch(`${GRPC_URL}/chat.ChatService/StreamMessages`, {
    method: "POST",
    headers: grpcHeaders(token),
    body: frameMessage(new Uint8Array(0)), // StreamRequest {} has no fields
    signal,
  });
  if (!res.ok) {
    throw new Error(`gRPC-Web stream failed: HTTP ${res.status}`);
  }
  onOpen?.();

  const reader = res.body.getReader();
  let buffer = new Uint8Array(0);
  for (;;) {
    const { done, value } = await reader.read();
    if (done) return;
    buffer = concatBytes(buffer, value);
    const { frames, rest } = extractFrames(buffer);
    buffer = rest;
    for (const frame of frames) {
      if (frame.isTrailer) {
        readTrailerStatus(frame.payload);
        return;
      }
      onMessage(decodeMessage(frame.payload));
    }
  }
}
