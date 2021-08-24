.PHONY: compile assets client

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
GO_BINDATA := $(GOPATH)/bin/go-bindata



configsaver.pb.go: proto/configsaver.proto
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/configsaver.proto

# This is a "phony" target - an alias for the above command, so "make compile"
# still works.
compile: configsaver.pb.go


serve:
	go run server/server.go

client:
	go run client/main.go