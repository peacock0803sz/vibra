import { createConnectTransport } from "@connectrpc/connect-web";

const DEFAULT_BASE_URL = "http://127.0.0.1:3001";

/**
 * Create a ConnectRPC transport for browser use.
 * Defaults to the local backend address in development.
 */
export function createTransport(baseUrl?: string) {
  return createConnectTransport({
    baseUrl: baseUrl ?? DEFAULT_BASE_URL,
  });
}
