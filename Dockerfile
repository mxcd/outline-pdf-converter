FROM golang:1.22-alpine3.18 AS builder
RUN apk add --no-cache git

WORKDIR /usr/src
COPY go.mod /usr/src/go.mod
COPY go.sum /usr/src/go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY internal internal

RUN go build -o outline-pdf-converter -ldflags="-s -w" cmd/server/main.go 

FROM alpine:3.18
WORKDIR /usr/bin
COPY --from=builder /usr/src/outline-pdf-converter .
ENTRYPOINT ["/usr/bin/outline-pdf-converter"]