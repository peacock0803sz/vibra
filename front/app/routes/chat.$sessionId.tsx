import { useParams } from "react-router";

import { ChatWindow } from "~/components/chat/ChatWindow";

export function meta() {
  return [{ title: "Chat - vibra" }, { name: "description", content: "AI agent streaming chat" }];
}

export default function ChatRoute() {
  const { sessionId } = useParams();

  return (
    <div className="flex h-dvh flex-col bg-white text-gray-900 dark:bg-gray-900 dark:text-gray-100">
      <header className="flex shrink-0 items-center justify-between border-b border-gray-200 px-4 py-3 dark:border-gray-700">
        <h1 className="text-lg font-semibold">vibra</h1>
        <span className="font-mono text-xs text-gray-400">{sessionId?.slice(0, 8)}</span>
      </header>
      <ChatWindow sessionId={sessionId} />
    </div>
  );
}
