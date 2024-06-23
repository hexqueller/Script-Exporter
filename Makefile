.PHONY: run
run:
	go run cmd/exporter/main.go

.PHONY: build
build:
	go build -o exporter cmd/exporter/main.go