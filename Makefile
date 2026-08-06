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
