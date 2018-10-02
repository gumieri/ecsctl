GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=ecsctl

all: deps build
deps:
	$(GOGET) ./...
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
zsh-completion:
	./${BINARY_NAME} completion zsh > _${BINARY_NAME}
zsh-install:
	mv ./${BINARY_NAME} /usr/local/bin
	mv ./_${BINARY_NAME} /usr/local/share/zsh/site-functions
zsh-uninstall:
	rm -f /usr/local/bin/${BINARY_NAME} 
	rm /usr/local/share/zsh/site-functions/_${BINARY_NAME} 
