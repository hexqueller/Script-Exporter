.SILENT:

build:
	go mod download && CGO_ENABLED=0 GOOS=linux go build -o exporter cmd/exporter/main.go

run: build
	./exporter -d

docker: build
	docker build . -t script-exporter && docker run -p 9105:9105 script-exporter
