import { renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { EnvironmentInfo, StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { AgentType } from "@gen/vibra/agent/v1/agent_pb";

import { useStatusLine } from "./useStatusLine";

// テスト用の EnvironmentInfo イベントファクトリ
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
  it("初期状態 (events が空) で全項目 '-' を返す", () => {
    const { result } = renderHook(() => useStatusLine([]));
    expect(result.current.hostname).toBe("-");
    expect(result.current.repository).toBe("-");
    expect(result.current.branch).toBe("-");
    expect(result.current.agentType).toBe("-");
    expect(result.current.modelName).toBe("-");
  });

  it("EnvironmentInfo イベントから全フィールドを取得する", () => {
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

  it("空文字フィールドを '-' として表示する", () => {
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

  it("複数のイベントがある場合、最後の EnvironmentInfo を使う", () => {
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

  it("EnvironmentInfo 以外のイベントのみの場合、'-' を返す", () => {
    const events = [makeTextEvent("hello"), makeTextEvent("world")];
    const { result } = renderHook(() => useStatusLine(events));
    expect(result.current.hostname).toBe("-");
    expect(result.current.repository).toBe("-");
  });

  it("エージェント種類を正しくラベル変換する (Codex, Gemini)", () => {
    const { result: codex } = renderHook(() =>
      useStatusLine([makeEnvEvent({ agent: AgentType.CODEX })]),
    );
    expect(codex.current.agentType).toBe("Codex");

    const { result: gemini } = renderHook(() =>
      useStatusLine([makeEnvEvent({ agent: AgentType.GEMINI })]),
    );
    expect(gemini.current.agentType).toBe("Gemini");
  });

  it("FR-011: 新しいイベントでフィールドが空文字の場合、直前の値を維持する", () => {
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

    // Git 管理外になった場合: repository/branch が空
    rerender({
      events: [
        ...initialEvents,
        makeEnvEvent({ hostname: "myhost", repository: "", branch: "", modelName: "claude-code" }),
      ],
    });

    // 直前の非空値が維持される (FR-011)
    expect(result.current.repository).toBe("owner/repo");
    expect(result.current.branch).toBe("main");
  });

  it("FR-011: 非空値に更新された場合は新しい値を表示する", () => {
    const initialEvents = [
      makeEnvEvent({ repository: "old/repo", branch: "old-branch" }),
    ];
    const { result, rerender } = renderHook(
      ({ events }: { events: StreamEvent[] }) => useStatusLine(events),
      { initialProps: { events: initialEvents } },
    );

    expect(result.current.repository).toBe("old/repo");

    rerender({
      events: [
        ...initialEvents,
        makeEnvEvent({ repository: "new/repo", branch: "new-branch" }),
      ],
    });

    expect(result.current.repository).toBe("new/repo");
    expect(result.current.branch).toBe("new-branch");
  });
});
