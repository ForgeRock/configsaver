.PHONY: compile assets

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
GO_BINDATA := $(GOPATH)/bin/go-bindata



app.pb.go: proto/app.proto
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/app.proto

# This is a "phony" target - an alias for the above command, so "make compile"
# still works.
compile: app.pb.go


serve:
	go run server/main.go