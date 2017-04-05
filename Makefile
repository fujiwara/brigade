all: brigade.go cmd/brigade/main.go
	go build ./cmd/brigade/

test:
	go test -v .
