import type { StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { useState, useCallback, useRef } from "react";

export type StreamStatus = "idle" | "streaming" | "done" | "error";

export interface UseStreamResult {
  events: StreamEvent[];
  status: StreamStatus;
  error: Error | null;
  /** 新しいイベントの非同期イテラブルを消費する */
  consume: (iterable: AsyncIterable<{ event?: StreamEvent }>) => Promise<void>;
  /** ストリーミングを中断する */
  abort: () => void;
  /** 状態をリセットする */
  reset: () => void;
}

/**
 * サーバーストリーミングRPCのイベント消費を管理するフック。
 * connect-esのAsyncIterableを受け取り、React状態に変換する。
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
