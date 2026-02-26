import type { EnvironmentInfo, StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { AgentType } from "@gen/vibra/agent/v1/agent_pb";
import { renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { useStatusLine } from "./useStatusLine";

// Factory for test EnvironmentInfo events
function makeEnvEvent(partial: Partial<EnvironmentInfo>): StreamEvent {
  return {
    sequence: 1n,
    messageId: "",
    payload: {
      case: "environment",
      value: {
        hostname: "",
        repository: "",
        branch: "",
        agent: AgentType.CLAUDE,
        modelName: "",
        ...partial,
      } as EnvironmentInfo,
    },
  } as StreamEvent;
}

function makeTextEvent(text: string): StreamEvent {
  return {
    sequence: 2n,
    messageId: "",
    payload: { case: "text", value: text },
  } as StreamEvent;
}

describe("useStatusLine", () => {
  it("returns '-' for all fields in initial state (empty events)", () => {
    const { result } = renderHook(() => useStatusLine([]));
    expect(result.current.hostname).toBe("-");
    expect(result.current.repository).toBe("-");
    expect(result.current.branch).toBe("-");
    expect(result.current.agentType).toBe("-");
    expect(result.current.modelName).toBe("-");
  });

  it("extracts all fields from EnvironmentInfo event", () => {
    const events = [
      makeEnvEvent({
        hostname: "myhost",
        repository: "owner/repo",
        branch: "main",
        agent: AgentType.CLAUDE,
        modelName: "claude-code",
      }),
    ];
    const { result } = renderHook(() => useStatusLine(events));
    expect(result.current.hostname).toBe("myhost");
    expect(result.current.repository).toBe("owner/repo");
    expect(result.current.branch).toBe("main");
    expect(result.current.agentType).toBe("Claude");
    expect(result.current.modelName).toBe("claude-code");
  });

  it("displays empty string fields as '-'", () => {
    const events = [
      makeEnvEvent({
        hostname: "myhost",
        repository: "",
        branch: "",
        agent: AgentType.CLAUDE,
        modelName: "claude-code",
      }),
    ];
    const { result } = renderHook(() => useStatusLine(events));
    expect(result.current.hostname).toBe("myhost");
    expect(result.current.repository).toBe("-");
    expect(result.current.branch).toBe("-");
    expect(result.current.modelName).toBe("claude-code");
  });

  it("uses the last EnvironmentInfo when multiple events exist", () => {
    const events = [
      makeEnvEvent({ hostname: "host1", repository: "r1", branch: "b1", modelName: "m1" }),
      makeTextEvent("hello"),
      makeEnvEvent({ hostname: "host2", repository: "r2", branch: "b2", modelName: "m2" }),
    ];
    const { result } = renderHook(() => useStatusLine(events));
    expect(result.current.hostname).toBe("host2");
    expect(result.current.repository).toBe("r2");
    expect(result.current.branch).toBe("b2");
    expect(result.current.modelName).toBe("m2");
  });

  it("returns '-' when only non-EnvironmentInfo events exist", () => {
    const events = [makeTextEvent("hello"), makeTextEvent("world")];
    const { result } = renderHook(() => useStatusLine(events));
    expect(result.current.hostname).toBe("-");
    expect(result.current.repository).toBe("-");
  });

  it("correctly converts agent types to labels (Codex, Gemini)", () => {
    const { result: codex } = renderHook(() =>
      useStatusLine([makeEnvEvent({ agent: AgentType.CODEX })]),
    );
    expect(codex.current.agentType).toBe("Codex");

    const { result: gemini } = renderHook(() =>
      useStatusLine([makeEnvEvent({ agent: AgentType.GEMINI })]),
    );
    expect(gemini.current.agentType).toBe("Gemini");
  });

  it("FR-011: preserves previous value when new event field is empty", () => {
    const initialEvents = [
      makeEnvEvent({
        hostname: "myhost",
        repository: "owner/repo",
        branch: "main",
        modelName: "claude-code",
      }),
    ];
    const { result, rerender } = renderHook(
      ({ events }: { events: StreamEvent[] }) => useStatusLine(events),
      { initialProps: { events: initialEvents } },
    );

    expect(result.current.repository).toBe("owner/repo");
    expect(result.current.branch).toBe("main");

    // Outside git management: repository/branch become empty
    rerender({
      events: [
        ...initialEvents,
        makeEnvEvent({ hostname: "myhost", repository: "", branch: "", modelName: "claude-code" }),
      ],
    });

    // Previous non-empty values are preserved (FR-011)
    expect(result.current.repository).toBe("owner/repo");
    expect(result.current.branch).toBe("main");
  });

  it("FR-011: displays new value when updated to non-empty", () => {
    const initialEvents = [makeEnvEvent({ repository: "old/repo", branch: "old-branch" })];
    const { result, rerender } = renderHook(
      ({ events }: { events: StreamEvent[] }) => useStatusLine(events),
      { initialProps: { events: initialEvents } },
    );

    expect(result.current.repository).toBe("old/repo");

    rerender({
      events: [...initialEvents, makeEnvEvent({ repository: "new/repo", branch: "new-branch" })],
    });

    expect(result.current.repository).toBe("new/repo");
    expect(result.current.branch).toBe("new-branch");
  });
});
