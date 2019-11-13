FROM golang:1 AS build

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

ARG VERSION

ENV CGO_ENABLED=0

RUN go build -ldflags "-X github.com/gumieri/ecsctl/cmd.VERSION=${VERSION}"

FROM scratch

COPY --from=build ["/ecsctl", "./"]

ENTRYPOINT ["/ecsctl"]
