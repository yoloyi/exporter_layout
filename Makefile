PROJECT_LINUX:=main


.PHONY: build-linux clean all

all:
	make build-linux

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -ldflags "-s -w" -o ./deploy/$(PROJECT_LINUX) ./scan.go

clean:
	rm -f $(PROJECT_LINUX)

