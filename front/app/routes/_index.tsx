import { useNavigate } from "react-router";
import { useCallback } from "react";

export function meta() {
  return [
    { title: "vibra" },
    { name: "description", content: "Vibe coding from any device" },
  ];
}

export default function IndexRoute() {
  const navigate = useNavigate();

  const handleNewSession = useCallback(() => {
    const sessionId = crypto.randomUUID();
    navigate(`/chat/${sessionId}`);
  }, [navigate]);

  return (
    <div className="h-dvh flex flex-col items-center justify-center bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 px-4">
      <div className="max-w-md w-full text-center space-y-8">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight">vibra</h1>
          <p className="text-gray-500 dark:text-gray-400">
            Vibe coding from any device
          </p>
        </div>

        <button
          onClick={handleNewSession}
          className="w-full rounded-lg bg-blue-600 px-6 py-3 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
        >
          New Session
        </button>
      </div>
    </div>
  );
}
