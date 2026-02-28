import { client } from "jsr:@probitas/probitas@^0";

// ConnectRPC back-end HTTP client factory targeting 127.0.0.1:3001.
// Use as a .resource() factory in Probitas scenarios.
export const backendClient = () =>
  client.http.createHttpClient({ url: "http://127.0.0.1:3001" });

// React Router front-end HTTP client factory targeting 127.0.0.1:3000.
// Use for server-side route testing (POST / for session creation, GET / for home page).
export const frontendClient = () =>
  client.http.createHttpClient({ url: "http://127.0.0.1:3000" });
