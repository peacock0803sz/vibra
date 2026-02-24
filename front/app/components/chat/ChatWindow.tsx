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
    <div className="flex flex-col h-full">
      <MessageList events={events} />

      {status === "streaming" && (
        <div className="px-4 py-1 text-xs text-gray-400 flex items-center gap-2">
          <span className="inline-block w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
          Agent is responding...
        </div>
      )}

      {error && (
        <div className="mx-4 mb-2 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-4 py-3 text-sm text-red-700 dark:text-red-300">
          <p>{error.message}</p>
        </div>
      )}

      <PromptInput
        onSubmit={handleSubmit}
        disabled={status === "streaming"}
      />
    </div>
  );
}
