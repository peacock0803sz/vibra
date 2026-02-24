import { useParams } from "react-router";
import { ChatWindow } from "~/components/chat/ChatWindow";

export function meta() {
  return [
    { title: "Chat - vibra" },
    { name: "description", content: "AI agent streaming chat" },
  ];
}

export default function ChatRoute() {
  const { sessionId } = useParams();

  return (
    <div className="h-dvh flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <header className="shrink-0 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
        <h1 className="text-lg font-semibold">vibra</h1>
        <span className="text-xs text-gray-400 font-mono">
          {sessionId?.slice(0, 8)}
        </span>
      </header>
      <ChatWindow sessionId={sessionId} />
    </div>
  );
}
