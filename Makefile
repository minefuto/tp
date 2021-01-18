NAME := tp
VERSION := $(shell git describe --tags --abbrev=0)
COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS := -s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'

all: build

test:
	@go test -v

clean:
	@go clean
	@rm -rf ./bin

build:
	@go build -ldflags "$(LDFLAGS)" -o ./bin/$(NAME)

run: build
	@./bin/$(NAME)

install:
	@go install -ldflags "$(LDFLAGS)"

uninstall:
	@rm $(GOBIN)/tp
