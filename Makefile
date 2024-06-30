.SILENT:

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o exporter cmd/exporter/main.go

run: build
	./exporter -d