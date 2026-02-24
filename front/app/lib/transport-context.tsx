import type { Transport } from "@connectrpc/connect";
import { createContext, useContext, useMemo } from "react";

import { createTransport } from "./connect-client";

const TransportContext = createContext<Transport | null>(null);

/**
 * Provides a ConnectRPC transport to child components.
 * Wrap the app root to make useTransport() available everywhere.
 */
export function TransportProvider({
  baseUrl,
  children,
}: {
  baseUrl?: string;
  children: React.ReactNode;
}) {
  const transport = useMemo(() => createTransport(baseUrl), [baseUrl]);
  return <TransportContext value={transport}>{children}</TransportContext>;
}

/**
 * Access the ConnectRPC transport from any component.
 * Must be used within a TransportProvider.
 */
export function useTransport(): Transport {
  const transport = useContext(TransportContext);
  if (!transport) {
    throw new Error("useTransport must be used within a TransportProvider");
  }
  return transport;
}
