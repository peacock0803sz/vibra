import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/_index.tsx"),
  route("chat/:sessionId", "routes/chat.$sessionId.tsx"),
] satisfies RouteConfig;
