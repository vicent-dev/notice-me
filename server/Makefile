run:
	go run ./cmd/server/main.go

test:
	go test -v ./... -cover

build:
	go build ./cmd/server/main.go

build-restart:
	go build ./cmd/server/main.go && systemctl restart notice-me
