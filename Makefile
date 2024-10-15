DOCKER_REPO=127.0.0.1:5000
IMAGE_TAG:=$(shell date +%s)

.PHONY: proto
	protoc --go_out=./client --go-grpc_out=./client ./message.proto
	protoc --go_out=./server --go-grpc_out=./server ./message.proto

.PHONY: client
client: proto
	cd client && CGO_ENABLED=0 go build .

.PHONY: server
server: proto
	cd server && CGO_ENABLED=0 go build .

.PHONY: pkg_client
pkg_client: client
	docker build -f ./client/Dockerfile -t ${DOCKER_REPO}/grpc-client:${IMAGE_TAG} ./client

.PHONY: pkg_server
pkg_server: server
	docker build -f ./server/Dockerfile -t ${DOCKER_REPO}/grpc-server:${IMAGE_TAG} ./server

.PHONY: pkg
pkg: pkg_client pkg_server
