.PHONY: build
build:
	go build -o dist/tfdiff ./cmd/tfdiff

.PHONY: install
install:
	go install github.com/takaishi/tfdiff/cmd/tfdiff

.PHONY: test
test:
	go test -race ./...
