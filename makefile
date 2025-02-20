all:
	CGO_ENABLED=0 go build -v
start:
	./bytepurr
clean:
	go fmt ./...