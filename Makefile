.PHONY: compile assets client docker

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
GO_BINDATA := $(GOPATH)/bin/go-bindata



configsaver.pb.go: proto/configsaver.proto
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/configsaver.proto

# This is a "phony" target - an alias for the above command, so "make compile"
# still works.
compile: configsaver.pb.go


serve:
	GIT_REPO="git@github.com:wstrange/forgeops.git" GIT_SSH_PATH=tmp/ssh go run server/config_server.go

client:
	CONFIG_DIR=tmp/client go run client/config_client.go

client_sync:
	CONFIG_DIR=tmp/client go run client/config_client.go 5

docker:
	docker build -t gcr.io/forgeops-public/config_client:dev  -f client/Dockerfile  .
	docker build -t gcr.io/forgeops-public/config_server:dev  -f server/Dockerfile .
	docker push gcr.io/forgeops-public/config_client:dev
	docker push gcr.io/forgeops-public/config_server:dev

cb:
	gcloud --project forgeops-public builds submit