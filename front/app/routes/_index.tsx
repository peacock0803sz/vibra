import { Form, redirect } from "react-router";

export function meta() {
  return [{ title: "vibra" }, { name: "description", content: "Vibe coding from any device" }];
}

export async function action() {
  const sessionId = crypto.randomUUID();
  return redirect(`/chat/${sessionId}`);
}

export default function IndexRoute() {
  return (
    <div className="flex h-dvh flex-col items-center justify-center bg-white px-4 text-gray-900 dark:bg-gray-900 dark:text-gray-100">
      <div className="w-full max-w-md space-y-8 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight">vibra</h1>
          <p className="text-gray-500 dark:text-gray-400">Vibe coding from any device</p>
        </div>

        <Form method="post">
          <button
            type="submit"
            className="w-full rounded-lg bg-blue-600 px-6 py-3 text-base font-medium text-white transition-colors hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
          >
            New Session
          </button>
        </Form>
      </div>
    </div>
  );
}
