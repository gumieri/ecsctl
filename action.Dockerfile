FROM golang:1 AS build

RUN apk update && apk add --no-cache ca-certificates

WORKDIR $GOPATH/src/gumieri/ecsctl

COPY go.mod go.sum ./

RUN go mod download

COPY ./*.go ./
COPY ./cmd/*.go ./cmd/

ARG VERSION

RUN CGO_ENABLED=0 go build -ldflags "-w -s -X github.com/gumieri/ecsctl/cmd.Version=${VERSION}" -o /usr/local/bin/ecsctl

ENTRYPOINT ["/usr/local/bin/ecsctl"]
