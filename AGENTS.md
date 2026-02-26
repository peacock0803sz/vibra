# vibra Development Guidelines

See [README.md](README.md) for project overview, architecture, and tech stack.

## Commands

```bash
# Proto
buf lint proto/ && buf generate

# Backend
cd back && go vet ./... && go test -race ./...

# Frontend
cd front && pnpm lint && pnpm format:check && pnpm typecheck && pnpm test
```

## Code Style

- Go: standard library conventions, `gofmt`
- TypeScript: oxlint + oxfmt (Rust-based, zero-config tools)
- Proto: `buf lint` enforced
- All documentation must be in English
- Commit messages: use emoji prefixes defined in [.gitmessage](.gitmessage)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
