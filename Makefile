.PHONY: build web-build go-build test test-go test-web dev-server dev-web install uninstall clean

web-build:
	cd web && bun install && bun run build

go-build: web-build
	go build -o bin/mission-control ./cmd/mission-control

build: go-build

test-go:
	go test -race ./...

test-web:
	cd web && bun test

test: test-go test-web

dev-server:
	go run ./cmd/mission-control serve --port 4317

dev-web:
	cd web && bun run dev

install: build
	./bin/mission-control install

uninstall:
	./bin/mission-control uninstall

clean:
	rm -rf bin web/dist/* web/node_modules
