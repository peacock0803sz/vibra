import { expect, scenario } from "jsr:@probitas/probitas@^0";
import { frontendClient } from "../helpers/connect-client.ts";

// US3: Verifies the React Router session creation flow.
// POST / creates a session and redirects to /chat/:sessionId.
// GET / renders the home page.
// Uses front-end HTTP client only — does not require Docker runtime.
export default scenario("Session Creation", { tags: ["session", "smoke"] })
  .resource("frontend", frontendClient)
  .step("GET / renders home page", async (ctx) => {
    const response = await ctx.resources.frontend.get("/");
    expect(response).toHaveStatus(200);
  })
  .step("POST / creates session and redirects to /chat/:sessionId", async (ctx) => {
    // React Router action handler creates a session UUID and returns a redirect.
    const response = await ctx.resources.frontend.post("/?index", {
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: "",
      redirect: "manual",
    });
    expect(response).toHaveStatus(302);
    const location = response.headers.get("location") ?? "";
    expect(location).toMatch(/^\/chat\/[0-9a-f-]{36}$/);
    // Pass extracted session ID to the next step via ctx.previous.
    return { sessionId: location.replace("/chat/", "") };
  })
  .step("GET /chat/:sessionId returns 200", async (ctx) => {
    const { sessionId } = ctx.previous as { sessionId: string };
    const response = await ctx.resources.frontend.get(`/chat/${sessionId}`);
    expect(response).toHaveStatus(200);
  })
  .build();
