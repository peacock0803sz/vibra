import { useCallback } from "react";

import { useAgent } from "~/lib/hooks/useAgent";

import { MessageList } from "./MessageList";
import { PromptInput } from "./PromptInput";

interface ChatWindowProps {
  sessionId?: string;
  workingDirectory?: string;
}

export function ChatWindow({ sessionId, workingDirectory }: ChatWindowProps) {
  const { events, status, error, execute } = useAgent();

  const handleSubmit = useCallback(
    (prompt: string) => {
      execute({
        prompt,
        sessionId,
        workingDirectory,
      });
    },
    [execute, sessionId, workingDirectory],
  );

  return (
    <div className="flex h-full flex-col">
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
