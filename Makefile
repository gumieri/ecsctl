MODULE=github.com/gumieri/ecsctl
GOBUILD=go build -ldflags "-X ${MODULE}/cmd.Version=${VERSION}"

define RELEASE_BODY
Go to the [README](https://github.com/gumieri/ecsctl/blob/${VERSION}/README.md) to know how to use it.
If you are using Mac OS or Linux, you can install using the commands:
```bash
curl -L https://github.com/gumieri/ecsctl/releases/download/${VERSION}/ecsctl-`uname -s`-`uname -m` -o /usr/local/bin/ecsctl
chmod +x /usr/local/bin/ecsctl
```
If you already have an older version installed, just run:
```bash
ecsctl upgrade
```
endef
export RELEASE_BODY

all: deps build
deps:
	go get ./...
upgrade:
	go get -u ./...
	go mod tidy
build:
	$(GOBUILD)
install:
	go install
release-body:
	echo "$$RELEASE_BODY" > RELEASE.md
build-linux-64:
	GOOS=linux \
	GOARCH=amd64 \
	$(GOBUILD) -o release/ecsctl-Linux-x86_64
build-linux-86:
	GOOS=linux \
	GOARCH=386 \
	$(GOBUILD) -o release/ecsctl-Linux-i386
build-linux-aarch64:
	GOOS=linux \
	GOARCH=arm64 \
	$(GOBUILD) -o release/ecsctl-Linux-aarch64
build-macos-64:
	GOOS=darwin \
	GOARCH=amd64 \
	$(GOBUILD) -o release/ecsctl-Darwin-x86_64
build-macos-arm64:
	GOOS=darwin \
	GOARCH=arm64 \
	$(GOBUILD) -o release/ecsctl-Darwin-arm64
build-all: build-linux-64 build-linux-86 build-linux-arm5 build-linux-arm6 build-linux-arm7 build-linux-arm8 build-macos-64 build-macos-arm64
