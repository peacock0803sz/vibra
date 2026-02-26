import { AgentType, type StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { useMemo } from "react";

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

const DEFAULT_DATA: StatusLineData = {
  hostname: FALLBACK,
  repository: FALLBACK,
  branch: FALLBACK,
  agentType: FALLBACK,
  modelName: FALLBACK,
};

/**
 * Extracts the latest EnvironmentInfo from streaming events for the status line.
 * Initial state shows "-" for all fields. Empty string fields also show "-".
 * Non-empty field values from earlier events are preserved if a later event has empty values (FR-011).
 */
export function useStatusLine(events: StreamEvent[]): StatusLineData {
  return useMemo(() => {
    let result = DEFAULT_DATA;
    for (const event of events) {
      if (event.payload.case === "environment") {
        const env = event.payload.value;
        result = {
          hostname: env.hostname || result.hostname,
          repository: env.repository || result.repository,
          branch: env.branch || result.branch,
          agentType: AGENT_LABELS[env.agent] || result.agentType,
          modelName: env.modelName || result.modelName,
        };
      }
    }
    return result;
  }, [events]);
}
