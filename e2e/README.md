# E2E Tests

API-level end-to-end tests for vibra using [Probitas](https://github.com/probitas-test/probitas), a Deno/TypeScript scenario-based HTTP testing framework.

These tests exercise ConnectRPC endpoints and React Router server-side routes at the HTTP level. They complement unit tests (vitest) and future browser-level E2E (Playwright).

## Prerequisites

- Go 1.25+
- Node.js 22 + pnpm
- Probitas CLI
- Docker runtime (required for chat scenario — agent container execution)

### Install Probitas

```bash
# Nix (recommended, matches project tooling)
nix profile install github:probitas-test/probitas

# Install script
curl -fsSL https://raw.githubusercontent.com/probitas-test/probitas/main/install.sh | bash

# Homebrew
brew tap probitas-test/tap && brew install probitas
```

Verify: `probitas --version`

## Run Tests

### Single command (manages server lifecycle)

```bash
./e2e/run-e2e.sh
```

Starts the Go back-end and React Router front-end, runs all scenarios, then tears down both servers.

### Manual (servers already running)

```bash
cd e2e && probitas run
```

### Run by tag

```bash
cd e2e && probitas run -s tag:health
cd e2e && probitas run -s tag:smoke
cd e2e && probitas run -s tag:chat
```

## Write a New Scenario

Create a file in `e2e/scenarios/` with the `.probitas.ts` extension:

```typescript
import { scenario } from "jsr:@probitas/probitas@^0";
import { backendClient } from "../helpers/connect-client.ts";

export default scenario("My scenario", { tags: ["smoke"] })
  .resource("api", backendClient)
  .step("step description", async (ctx) => {
    const response = await ctx.resources.api.post(
      "/vibra.agent.v1.AgentService/GetNodeInfo",
      { body: {} },
    );
    expect(response).toHaveStatus(200);
    expect(response).toHaveJsonProperty("nodeId");
  })
  .build();
```

Run it immediately: `cd e2e && probitas run`

## Available Helpers

| File | Purpose |
|------|---------|
| `helpers/connect-client.ts` | `backendClient()` — ConnectRPC back-end at `127.0.0.1:3001`; `frontendClient()` — React Router front-end at `127.0.0.1:3000` |
| `helpers/stream-client.ts` | `callConnectStreaming()` — raw stream reader for ConnectRPC server-streaming endpoints (`Execute`, `ContinueSession`) |

### Passing data between steps

Use `return` from a step to pass data to the next step via `ctx.previous`:

```typescript
.step("create session", async (ctx) => {
  // ...
  return { sessionId: "..." };
})
.step("use session", async (ctx) => {
  const { sessionId } = ctx.previous as { sessionId: string };
})
```

## Environment Variables

Set automatically by `run-e2e.sh`. For manual runs, set these before starting your servers:

| Variable | Value | Purpose |
|----------|-------|---------|
| `VIBRA_LISTEN_ADDR` | `127.0.0.1:3001` | Back-end listen address |
| `VIBRA_CORS_ORIGIN` | `http://127.0.0.1:3000` | CORS origin for front-end |
| `VIBRA_DEV_USER` | `e2e-test-user` | Auth bypass for ConnectRPC interceptor |
| `VIBRA_ALLOWED_DIRS` | `/tmp/vibra-e2e` | Sandbox directory for agent execution |
| `VIBRA_DEFAULT_WORKDIR` | `/tmp/vibra-e2e` | Default working directory for agent execution |

**Note**: Chat scenario tests require `ANTHROPIC_API_KEY` (or `OPENAI_API_KEY` / `GEMINI_API_KEY` for other agents) to be set on the host. The sandbox forwards API keys from the host environment to agent containers.

## Scenarios

| File | Tags | Requires Docker |
|------|------|-----------------|
| `scenarios/health.probitas.ts` | `health`, `smoke` | No |
| `scenarios/session.probitas.ts` | `session`, `smoke` | No |
| `scenarios/chat.probitas.ts` | `chat`, `smoke`, `docker` | Yes |
