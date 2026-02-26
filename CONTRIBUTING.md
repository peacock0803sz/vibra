# Contributing to Vibra

Vibra is MIT-licensed and accepts contributions.

## Development Setup

The recommended approach uses Nix with direnv:

```bash
direnv allow   # or: nix develop
```

This gives you Go, Node.js, Buf CLI, and other tools. Without Nix, install them manually:

- Go 1.25+
- Node.js 22 + pnpm
- [Buf CLI](https://buf.build/docs/installation)
- Docker or Podman

## Building

Generate code from Protobuf definitions first:

```bash
buf generate
```

Then start each component:

```bash
# Backend
cd back && go run ./cmd/vibra/

# Frontend (separate terminal)
cd front && pnpm install && pnpm dev
```

## Code Style

- Go follows standard library conventions. `gofmt` handles formatting.
- TypeScript uses oxlint for linting and oxfmt for formatting. No extra config needed.
- Proto files must pass `buf lint`.
- Write documentation and comments in English.
- Use emoji prefixes in commit messages. See [.gitmessage](.gitmessage) for the full list.

## Checks

Run these before opening a PR:

```bash
# Proto
buf lint proto/

# Backend
cd back && go vet ./... && go test -race ./...

# Frontend
cd front && pnpm lint && pnpm format:check && pnpm typecheck && pnpm test
```

Or use the Makefile targets:

```bash
make buf-lint
make back-lint && make back-test
make front-lint && make front-typecheck && make front-test
```

## Submitting Changes

1. Fork the repo and create a branch from `main`.
2. Make your changes, keeping commits focused on a single concern.
3. Ensure all checks above pass.
4. Open a pull request against `main` with a clear description of what you changed and why.

## Project Layout

See the [README](README.md#architecture) for the directory structure.

## Reporting Issues

Open an issue on GitHub. Include steps to reproduce and any relevant logs or error messages.
