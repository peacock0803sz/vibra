import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { AgentService, AgentType, ExecuteRequestSchema } from "@gen/vibra/agent/v1/agent_pb";
import { useCallback, useMemo } from "react";

import { useTransport } from "../transport-context";
import { useStream } from "./useStream";
import type { UseStreamResult } from "./useStream";

export interface ExecuteParams {
  prompt: string;
  agent?: AgentType;
  sessionId?: string;
  workingDirectory?: string;
}

export interface UseAgentResult extends UseStreamResult {
  execute: (params: ExecuteParams) => Promise<void>;
}

/**
 * Hook to invoke the Execute RPC.
 * Uses useStream internally and delegates streaming state management.
 */
export function useAgent(): UseAgentResult {
  const transport = useTransport();
  const stream = useStream();

  const client = useMemo(() => createClient(AgentService, transport), [transport]);

  const execute = useCallback(
    async (params: ExecuteParams) => {
      const req = create(ExecuteRequestSchema, {
        agent: params.agent ?? AgentType.CLAUDE,
        prompt: params.prompt,
        sessionId: params.sessionId ?? "",
        workingDirectory: params.workingDirectory ?? "",
      });

      const responseStream = client.execute(req);
      await stream.consume(responseStream);
    },
    [client, stream],
  );

  return { ...stream, execute };
}
