import { createConnectTransport } from "@connectrpc/connect-web";
import type { Transport } from "@connectrpc/connect";

// Gateway service base URL. The desktop app talks ONLY to gateway-service —
// not to upstream services directly.
const DEFAULT_BASE_URL = "http://localhost:7000";

export function getBaseUrl(): string {
  const fromEnv = import.meta.env.VITE_GATEWAY_URL as string | undefined;
  return fromEnv && fromEnv.length > 0 ? fromEnv : DEFAULT_BASE_URL;
}

/**
 * ConnectRPC transport for the browser/Tauri renderer. Generated Connect-ES
 * service clients are constructed with `createClient(ServiceDesc, transport)`
 * once the protogen step lands.
 */
export const transport: Transport = createConnectTransport({
  baseUrl: getBaseUrl(),
  // We use Connect (not gRPC-Web) so we can stream over plain HTTP in dev.
  useBinaryFormat: false,
  // Allow long-lived streams (RunEvents, audit tail, flow generation).
  fetch: globalThis.fetch.bind(globalThis),
});

/**
 * The constructed Connect service clients live in ./clients.ts and import
 * `transport` from this file. Pages should import from ./clients only.
 *
 * Runtime note: the gateway currently serves gRPC; for browser clients to
 * actually reach it, see services/gateway/README.md for the two supported
 * options (add Connect-Go handlers or run a gRPC-Web proxy).
 */
