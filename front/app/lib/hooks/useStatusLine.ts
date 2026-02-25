import { AgentType, type StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { useRef } from "react";

export interface StatusLineData {
  hostname: string;
  repository: string;
  branch: string;
  agentType: string;
  modelName: string;
}

const FALLBACK = "-";

const AGENT_LABELS: Record<number, string> = {
  [AgentType.CLAUDE]: "Claude",
  [AgentType.CODEX]: "Codex",
  [AgentType.GEMINI]: "Gemini",
};

// ES2022互換: findLast 不可なため逆順走査で最後のEnvironmentInfoを取得
function findLatestEnvironment(events: StreamEvent[]) {
  return [...events].reverse().find((e) => e.payload.case === "environment");
}

/**
 * Extracts the latest EnvironmentInfo from streaming events for the status line.
 * Initial state shows "-" for all fields. Empty string fields also show "-".
 * Non-empty field values from previous events are preserved if the new event has empty values (FR-011).
 */
export function useStatusLine(events: StreamEvent[]): StatusLineData {
  const prevRef = useRef<StatusLineData>({
    hostname: FALLBACK,
    repository: FALLBACK,
    branch: FALLBACK,
    agentType: FALLBACK,
    modelName: FALLBACK,
  });

  const latestEvent = findLatestEnvironment(events);
  if (!latestEvent || latestEvent.payload.case !== "environment") {
    return prevRef.current;
  }

  const env = latestEvent.payload.value;

  // 空文字のフィールドは直前の値を維持 (FR-011)
  const next: StatusLineData = {
    hostname: env.hostname || prevRef.current.hostname,
    repository: env.repository || prevRef.current.repository,
    branch: env.branch || prevRef.current.branch,
    agentType:
      AGENT_LABELS[env.agent] || prevRef.current.agentType,
    modelName: env.modelName || prevRef.current.modelName,
  };

  prevRef.current = next;
  return next;
}
