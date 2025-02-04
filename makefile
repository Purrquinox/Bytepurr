all:
	CGO_ENABLED=0 go build -v
start:
	./popkat
clean:
	go fmt ./...