.PHONY: buf-generate buf-lint back-test back-lint front-test front-lint

buf-lint:
	buf lint

buf-generate:
	buf generate

back-test:
	cd back && go test -race ./...

back-lint:
	cd back && go vet ./...

front-test:
	cd front && pnpm test

front-lint:
	cd front && pnpm lint

front-typecheck:
	cd front && pnpm typecheck
