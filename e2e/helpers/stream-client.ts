// T010a streaming spike decision:
// Probitas's built-in HTTP client handles standard request/response patterns.
// ConnectRPC server-streaming uses the Connect protocol envelope format:
// each message is preceded by a 5-byte header (1 flag byte + 4-byte big-endian length).
// Probitas does not natively parse these enveloped messages, so this helper uses
// Deno's native fetch (available via Probitas's bundled Deno runtime) to read the
// raw stream and parse envelopes manually.

const BACKEND_URL = `http://${Deno.env.get("VIBRA_LISTEN_ADDR") ?? "127.0.0.1:3001"}`;

// End-stream message flag in the Connect protocol envelope.
const FLAG_END_STREAM = 0x02;

export interface StreamEvent {
  sequence?: number;
  messageId?: string;
  text?: string;
  session?: {
    sessionId: string;
    agent: number;
    costMicros: number;
    currencyCode: string;
    nodeId: string;
  };
  toolUse?: unknown;
  toolResult?: unknown;
  error?: { code: number; message: string };
}

export interface StreamResult {
  events: StreamEvent[];
  endStreamMetadata: Record<string, unknown>;
}

// callConnectStreaming sends a ConnectRPC server-streaming POST request and
// returns all parsed events. On connection failure, the error message identifies
// 127.0.0.1:3001 as unreachable to aid CI debugging.
export async function callConnectStreaming(
  path: string,
  body: Record<string, unknown>,
  timeoutMs = 10000,
): Promise<StreamResult> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  let response: Response;
  try {
    response = await fetch(`${BACKEND_URL}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/connect+json",
        "Connect-Protocol-Version": "1",
      },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
  } catch (err) {
    clearTimeout(timeoutId);
    throw new Error(
      `ConnectRPC streaming request failed (${BACKEND_URL} unreachable): ${err}`,
    );
  }

  if (!response.ok) {
    clearTimeout(timeoutId);
    const errorBody = await response.text();
    throw new Error(
      `ConnectRPC streaming request failed with status ${response.status}: ${errorBody}`,
    );
  }

  const events: StreamEvent[] = [];
  let endStreamMetadata: Record<string, unknown> = {};
  let receivedEndStream = false;

  try {
    if (!response.body) {
      throw new Error("No response body from streaming RPC");
    }

    const reader = response.body.getReader();
    let buffer = new Uint8Array(0);

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      // Append new bytes to the running buffer.
      const next = new Uint8Array(buffer.length + value.length);
      next.set(buffer);
      next.set(value, buffer.length);
      buffer = next;

      // Parse all complete envelopes: 5-byte header + message body.
      while (buffer.length >= 5) {
        const flags = buffer[0];
        const length =
          (buffer[1] << 24) | (buffer[2] << 16) | (buffer[3] << 8) | buffer[4];

        if (buffer.length < 5 + length) break; // incomplete envelope, wait for more data

        const messageBytes = buffer.slice(5, 5 + length);
        buffer = buffer.slice(5 + length);

        const parsed = JSON.parse(new TextDecoder().decode(messageBytes));

        if (flags & FLAG_END_STREAM) {
          endStreamMetadata = parsed;
          receivedEndStream = true;
        } else {
          // ExecuteResponse wraps the payload as { event: StreamEvent }.
          events.push(parsed.event ?? parsed);
        }
      }
    }
  } finally {
    clearTimeout(timeoutId);
  }

  if (!receivedEndStream) {
    throw new Error("Stream terminated without end-stream message; possible premature server disconnect");
  }

  return { events, endStreamMetadata };
}
