FROM golang:alpine AS build

RUN apk update && apk add --no-cache ca-certificates

WORKDIR $GOPATH/src/gumieri/ecsctl

COPY go.mod go.sum ./

RUN go mod download

COPY ./*.go ./
COPY ./cmd/*.go ./cmd/

ARG VERSION

RUN CGO_ENABLED=0 go build -ldflags "-w -s -X github.com/gumieri/ecsctl/cmd.Version=${VERSION}" -o /go/bin/ecsctl

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
USER nobody

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build ["/go/bin/ecsctl", "/usr/local/bin/"]

ENTRYPOINT ["/usr/local/bin/ecsctl"]
