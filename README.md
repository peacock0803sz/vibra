# Vibra

AI agent interaction platform for solo developers.
Operate multiple AI agents (Claude Code, Codex CLI, Gemini CLI) from any device over your personal Tailscale network.

## Features

- **Streaming Chat** - Real-time response streaming with ConnectRPC (server-sent events)
- **Multi-Agent Support** - Pluggable adapter architecture for Claude, Codex, Gemini
- **Container Isolation** - Each agent runs in a Docker container with configurable permission profiles (OBSERVE / VERIFY / EDIT / FULL)
- **Tailscale Auth** - Zero-config authentication via Tailscale identity headers
- **Session Persistence** - Resume conversations across devices
- **Mobile Responsive** - Touch-optimized UI for on-the-go usage

## Architecture

```text
proto/vibra/           # Protobuf definitions (single source of truth)
back/                  # Go + connect-go server
  cmd/vibra/           # Main binary
  internal/            # adapter/, server/, auth/, container/, sandbox/
  gen/                 # buf generated Go code
front/                 # React Router v7 + connect-es client
  app/                 # routes/, components/, lib/
  gen/                 # buf generated TypeScript code
images/                # Agent container Dockerfiles (claude, codex, gemini)
```

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| Backend | Go 1.23+, connect-go |
| Frontend | React Router v7, TypeScript, connect-es, TailwindCSS |
| Protocol | Protocol Buffers v3, Buf CLI |
| Container | Docker / Podman |
| Linting | oxlint, oxfmt (Rust-based) |
| Dev Environment | Nix flake + direnv |

## Development

### Prerequisites

- [Nix](https://nixos.org/) with flakes enabled (or manually install Go 1.23+, Node.js 22, Buf CLI, Docker)
- [Tailscale](https://tailscale.com/) connected on your network

### Setup

```bash
# Enter dev shell (installs Go, Node, Buf, Docker client)
direnv allow
# or: nix develop

# Generate code from proto definitions
buf lint proto/ && buf generate

# Start backend
cd back && go run ./cmd/vibra/

# Start frontend (in another terminal)
cd front && pnpm install && pnpm dev
```

### Testing

```bash
# Proto
buf lint proto/

# Backend
cd back && go vet ./... && go test -race ./...

# Frontend
cd front && pnpm lint && pnpm format:check && pnpm typecheck && pnpm test
```