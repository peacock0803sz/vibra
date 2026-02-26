import { expect, scenario } from "jsr:@probitas/probitas@^0";
import { frontendClient } from "../helpers/connect-client.ts";
import { callConnectStreaming } from "../helpers/stream-client.ts";

// US3: Verifies the core chat interaction flow through the full stack.
// Requires Docker runtime for agent container execution.
// Uses stream-client.ts for ConnectRPC server-streaming (T010a decision).
export default scenario("Chat Interaction", { tags: ["chat", "smoke"] })
  .resource("frontend", frontendClient)
  .step("create session via POST /", async (ctx) => {
    const response = await ctx.resources.frontend.post("/", {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: "",
      followRedirects: false,
    });
    expect(response).toHaveStatus(302);
    const location = response.headers.get("location") ?? "";
    expect(location).toMatch(/^\/chat\/[0-9a-f-]{36}$/);
    return { sessionId: location.replace("/chat/", "") };
  })
  .step("GET /chat/:sessionId returns 200", async (ctx) => {
    const { sessionId } = ctx.previous as { sessionId: string };
    const response = await ctx.resources.frontend.get(`/chat/${sessionId}`);
    expect(response).toHaveStatus(200);
    return ctx.previous;
  })
  .step("Execute returns text and session events", async (ctx) => {
    // On back-end timeout, error message identifies 127.0.0.1:3001 as unreachable.
    const { sessionId } = ctx.previous as { sessionId: string };
    const result = await callConnectStreaming(
      "/vibra.agent.v1.AgentService/Execute",
      {
        agent: 1, // AGENT_TYPE_CLAUDE
        prompt: "Reply with exactly: ok",
        sessionId,
        workingDirectory: "/tmp/vibra-e2e",
        permissionMode: 1, // PERMISSION_MODE_OBSERVE
      },
      10000,
    );
    // Verify structural contract only — do not assert on model-generated text content.
    const textEvents = result.events.filter((e) => e.text !== undefined);
    const sessionEvents = result.events.filter((e) => e.session !== undefined);
    expect(textEvents.length).toBeGreaterThan(0);
    expect(sessionEvents.length).toBeGreaterThan(0);
  })
  .build();
