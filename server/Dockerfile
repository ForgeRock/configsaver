FROM golang:1.17-bullseye as BUILD

WORKDIR /app

RUN apt-get update && apt-get install -y libgit2-dev

COPY  ./ ./

RUN go mod download

RUN go build server/config_server.go

##
## Deploy
##
# Note - we could use distroless here if we can figure out the right libgit2 shared libs to copy in.
# FROM gcr.io/distroless/base-debian11
# For now use debian slim. It also useful to have some git commands for troubleshooting.
FROM debian:bullseye-slim

RUN adduser --home "/home/forgerock" -uid 11111 --gid 0 forgerock --disabled-password --gecos "forgerock user" && \
    apt-get update && apt-get install -y libgit2-dev git bash

WORKDIR /

COPY --from=build /app/config_server /config_server

EXPOSE 50051

# This is used in distroless only
# USER nonroot:nonroot

USER forgerock

ENTRYPOINT ["/config_server"]
