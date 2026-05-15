# SDD: Task 06 — Makefile CLI Target

## Summary

Add a `build-cli` target to the Makefile that builds the CLI binary.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `Makefile` | modify | Add `build-cli` and optionally `run-cli` targets |

## Current State

```makefile
run:
	go run ./cmd/server/main.go

test:
	go test -v ./... -cover

build:
	go build ./cmd/server/main.go

build-restart:
	go build ./cmd/server/main.go && systemctl restart notice-me
```

## Detailed Changes

**After:**
```makefile
run:
	go run ./cmd/server/main.go

test:
	go test -v ./... -cover

build:
	go build ./cmd/server/main.go

build-cli:
	go build -o notice-me-cli ./cmd/cli/main.go

run-cli:
	go run ./cmd/cli/main.go

build-restart:
	go build ./cmd/server/main.go && systemctl restart notice-me
```

## Test Plan

No tests. Manual verification:
1. `make build-cli` — should produce a `notice-me-cli` binary
2. `./notice-me-cli` — should print key to stdout (requires DB)

## Edge Cases

- **Missing CLI cmd**: If `cmd/cli/main.go` doesn't compile, `make build-cli` will show the Go compiler error. Task 1 ensures it compiles.
- **Binary name collision**: `notice-me-cli` is a unique name. No existing binary uses this name.
