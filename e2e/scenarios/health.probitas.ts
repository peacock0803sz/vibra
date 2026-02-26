import { expect, scenario } from "jsr:@probitas/probitas@^0";
import { backendClient } from "../helpers/connect-client.ts";

// US1: Validates that the ConnectRPC back-end is healthy and agents are available.
// Uses only unary RPCs — does not require Docker runtime.
export default scenario("Health Check", { tags: ["health", "smoke"] })
  .resource("api", backendClient)
  .step("GetNodeInfo returns node_id and status", async (ctx) => {
    const response = await ctx.resources.api.post(
      "/vibra.agent.v1.AgentService/GetNodeInfo",
      { body: {} },
    );
    expect(response).toHaveStatus(200);
    expect(response).toHaveJsonProperty("nodeId");
    expect(response).toHaveJsonProperty("status");
  })
  .step("ListAgents returns non-empty agent list", async (ctx) => {
    const response = await ctx.resources.api.post(
      "/vibra.agent.v1.AgentService/ListAgents",
      { body: {} },
    );
    expect(response).toHaveStatus(200);
    expect(response).toHaveJsonProperty("agents");
    const data = await response.json();
    expect(data.agents.length).toBeGreaterThan(0);
  })
  .build();
