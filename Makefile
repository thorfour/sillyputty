.PHONY: all clean docker 

EXE="server"
REPO="quay.io/thorfour/sillyputty"

docker:
	mkdir -p ./bin/docker
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/docker/$(EXE) ./cmd/server
	cp /etc/ssl/certs/ca-certificates.crt ./bin/docker
	cp ./build/Dockerfile ./bin/docker
	docker build ./bin/docker -t $(REPO)
clean:
	rm -r ./bin
all: docker
