import type { StreamEvent } from "@gen/vibra/agent/v1/agent_pb";
import { useCallback, useEffect, useRef, useState } from "react";

import type { UseAgentResult } from "~/lib/hooks/useAgent";

import { MessageList } from "./MessageList";
import { PromptInput } from "./PromptInput";

const STORAGE_KEY = "vibra:workingDirectory";

interface ChatWindowProps extends UseAgentResult {
  sessionId?: string;
  workingDirectory?: string;
}

// Extract the Claude-returned session ID from stream events.
function extractClaudeSessionId(events: StreamEvent[]): string | undefined {
  for (const ev of events) {
    if (ev.payload.case === "session" && ev.payload.value.sessionId) {
      return ev.payload.value.sessionId;
    }
  }
  return undefined;
}

export function ChatWindow({ sessionId: _sessionId, workingDirectory, events, status, error, execute }: ChatWindowProps) {
  const [workDir, setWorkDir] = useState(workingDirectory ?? "");
  // Tracks the Claude-returned session ID for --resume.
  const claudeSessionRef = useRef<string | undefined>(undefined);

  // Update Claude session ID whenever events change.
  useEffect(() => {
    const id = extractClaudeSessionId(events);
    if (id) claudeSessionRef.current = id;
  }, [events]);

  // Restore last-used directory from localStorage on mount.
  useEffect(() => {
    if (!workingDirectory) {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) setWorkDir(saved);
    }
  }, [workingDirectory]);

  const handleWorkDirChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setWorkDir(value);
    localStorage.setItem(STORAGE_KEY, value);
  }, []);

  const handleSubmit = useCallback(
    (prompt: string) => {
      execute({
        prompt,
        sessionId: claudeSessionRef.current,
        workingDirectory: workDir,
      });
    },
    [execute, workDir],
  );

  return (
    <div className="flex h-full flex-col">
      <div className="border-b border-gray-100 px-4 py-2 dark:border-gray-800">
        <div className="mx-auto flex max-w-4xl items-center gap-2">
          <label
            htmlFor="workdir"
            className="shrink-0 text-xs font-medium text-gray-500 dark:text-gray-400"
          >
            Working Directory
          </label>
          <input
            id="workdir"
            type="text"
            value={workDir}
            onChange={handleWorkDirChange}
            placeholder="/path/to/project"
            className="min-w-0 flex-1 rounded border border-gray-200 bg-gray-50 px-2 py-1 font-mono text-xs text-gray-700 placeholder:text-gray-400 focus:border-blue-400 focus:ring-1 focus:ring-blue-400 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
          />
        </div>
      </div>

      <MessageList events={events} />

      {status === "streaming" && (
        <div className="flex items-center gap-2 px-4 py-1 text-xs text-gray-400">
          <span className="inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-blue-400" />
          Agent is responding...
        </div>
      )}

      {error && (
        <div className="mx-4 mb-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300">
          <p>{error.message}</p>
        </div>
      )}

      <PromptInput onSubmit={handleSubmit} disabled={status === "streaming"} />
    </div>
  );
}
