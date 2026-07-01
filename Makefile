.PHONY: build test bench clean

build:
	go build -o attributor ./cmd/attributor

test:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -f attributor
	rm -rf .out/
	go clean -testcache
