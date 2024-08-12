FROM golang:1 AS build

WORKDIR $GOPATH/src/gumieri/ecsctl

COPY go.mod go.sum ./

RUN go mod download

COPY ./*.go ./
COPY ./cmd/*.go ./cmd/

ARG VERSION

RUN CGO_ENABLED=0 go build -ldflags "-w -s -X github.com/gumieri/ecsctl/cmd.Version=${VERSION}" -o /usr/local/bin/ecsctl

COPY ./action-entrypoint.sh /usr/local/bin/action-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/action-entrypoint.sh"]
