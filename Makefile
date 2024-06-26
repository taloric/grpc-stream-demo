DOCKER_REPO=127.0.0.1:5000
IMAGE_TAG:=$(shell date +%s)

.PHONY: client
client:
	protoc --go_out=./client --go-grpc_out=./client ./message.proto

.PHONY: server
server:
	protoc --go_out=./server --go-grpc_out=./server ./message.proto

.PHONY: build_client
build_client:
	cd client && CGO_ENABLED=0 go build .

.PHONY: build_server
build_server:
	cd server && CGO_ENABLED=0 go build .

.PHONY: package_server
package_server:
	docker build -f ./server/Dockerfile -t ${DOCKER_REPO}/grpc-server:${IMAGE_TAG} ./server

.PHONY: package_client
package_client:
	docker build -f ./client/Dockerfile -t ${DOCKER_REPO}/grpc-client:${IMAGE_TAG} ./client

.PHONY: all
all: build_client build_server package_server package_client