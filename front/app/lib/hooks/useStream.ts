import type { StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { useState, useCallback, useRef } from "react";

export type StreamStatus = "idle" | "streaming" | "done" | "error";

export interface UseStreamResult {
  events: StreamEvent[];
  status: StreamStatus;
  error: Error | null;
  /** Consumes an async iterable of new events */
  consume: (iterable: AsyncIterable<{ event?: StreamEvent }>) => Promise<void>;
  /** Aborts the stream consumption */
  abort: () => void;
  /** Resets the stream state */
  reset: () => void;
}

/**
 * Hook to manage consumption of server-streaming RPC events.
 * Takes an AsyncIterable from connect-es and translates it into React state.
 */
export function useStream(): UseStreamResult {
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const [status, setStatus] = useState<StreamStatus>("idle");
  const [error, setError] = useState<Error | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const abort = useCallback(() => {
    abortRef.current?.abort();
    abortRef.current = null;
  }, []);

  const reset = useCallback(() => {
    abort();
    setEvents([]);
    setStatus("idle");
    setError(null);
  }, [abort]);

  const consume = useCallback(
    async (iterable: AsyncIterable<{ event?: StreamEvent }>) => {
      abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setEvents([]);
      setStatus("streaming");
      setError(null);

      try {
        for await (const msg of iterable) {
          if (controller.signal.aborted) break;
          if (msg.event) {
            setEvents((prev) => [...prev, msg.event!]);
          }
        }
        if (!controller.signal.aborted) {
          setStatus("done");
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setError(err instanceof Error ? err : new Error(String(err)));
          setStatus("error");
        }
      }
    },
    [abort],
  );

  return { events, status, error, consume, abort, reset };
}
