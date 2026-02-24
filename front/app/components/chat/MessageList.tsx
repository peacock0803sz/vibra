import type { StreamEvent } from "@gen/vibra/agent/v1/agent_pb";

interface MessageListProps {
  events: StreamEvent[];
}

export function MessageList({ events }: MessageListProps) {
  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-3">
      {events.length === 0 && (
        <p className="text-center text-gray-400 mt-8">
          Send a prompt to start a conversation.
        </p>
      )}
      {events.map((ev, i) => (
        <MessageItem key={i} event={ev} />
      ))}
    </div>
  );
}

function MessageItem({ event }: { event: StreamEvent }) {
  switch (event.payload.case) {
    case "text":
      return <TextMessage text={event.payload.value} />;
    case "toolUse":
      return (
        <ToolUseMessage toolName={event.payload.value.toolName} />
      );
    case "toolResult":
      return (
        <ToolResultMessage
          toolName={event.payload.value.toolName}
          success={event.payload.value.success}
          output={event.payload.value.output}
          language={event.payload.value.language}
        />
      );
    case "session":
      return (
        <SessionMessage
          sessionId={event.payload.value.sessionId}
          costMicros={event.payload.value.costMicros}
        />
      );
    case "error":
      return <ErrorMessage message={event.payload.value.message} />;
    default:
      return null;
  }
}

function TextMessage({ text }: { text: string }) {
  return (
    <div className="rounded-lg bg-gray-50 dark:bg-gray-800 px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap break-words">
      {text}
    </div>
  );
}

function ToolUseMessage({ toolName }: { toolName: string }) {
  return (
    <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 px-4 py-1">
      <span className="inline-block w-2 h-2 rounded-full bg-blue-400 animate-pulse" />
      Using tool: <code className="font-mono bg-gray-100 dark:bg-gray-700 px-1 rounded">{toolName}</code>
    </div>
  );
}

function ToolResultMessage({
  toolName,
  success,
  output,
  language,
}: {
  toolName: string;
  success: boolean;
  output: string;
  language: string;
}) {
  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
      <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 dark:bg-gray-800 text-xs">
        <span className={success ? "text-green-600" : "text-red-500"}>
          {success ? "✓" : "✗"}
        </span>
        <code className="font-mono">{toolName}</code>
        {language && (
          <span className="ml-auto text-gray-400">{language}</span>
        )}
      </div>
      {output && (
        <pre className="p-3 text-xs font-mono overflow-x-auto bg-gray-50 dark:bg-gray-900 whitespace-pre-wrap break-words">
          <code>{output}</code>
        </pre>
      )}
    </div>
  );
}

function SessionMessage({
  sessionId,
  costMicros,
}: {
  sessionId: string;
  costMicros: bigint;
}) {
  const costUSD = Number(costMicros) / 1_000_000;
  return (
    <div className="text-xs text-gray-400 text-center py-1">
      Session: {sessionId.slice(0, 12)}...
      {costMicros > 0n && ` · $${costUSD.toFixed(4)}`}
    </div>
  );
}

function ErrorMessage({ message }: { message: string }) {
  return (
    <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 px-4 py-3 text-sm text-red-700 dark:text-red-300">
      {message}
    </div>
  );
}
