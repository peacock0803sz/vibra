import type { StatusLineData } from "~/lib/hooks/useStatusLine";

interface StatusLineProps {
  data: StatusLineData;
}

interface StatusItemProps {
  label: string;
  value: string;
  truncate?: boolean;
}

function StatusItem({ label, value, truncate = false }: StatusItemProps) {
  return (
    <span className="flex items-center gap-1">
      <span className="text-gray-400 dark:text-gray-500">{label}</span>
      <span
        className={`font-mono text-gray-700 dark:text-gray-300 ${
          truncate ? "max-w-32 truncate" : ""
        }`}
        title={truncate ? value : undefined}
      >
        {value}
      </span>
    </span>
  );
}

export function StatusLine({ data }: StatusLineProps) {
  return (
    <div className="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-1 border-b border-gray-200 bg-gray-50 px-4 py-1.5 text-xs dark:border-gray-700 dark:bg-gray-800/50">
      <StatusItem label="host" value={data.hostname} />
      <StatusItem label="repo" value={data.repository} truncate />
      <StatusItem label="branch" value={data.branch} truncate />
      <StatusItem label="agent" value={data.agentType} />
      <StatusItem label="model" value={data.modelName} />
    </div>
  );
}
