all: build test

build:
	go get && go build -v -o "terraform-provider-softlayer-${GIMME_OS}-${GIMME_ARCH}${EXT}"

test:
ifeq ($(RUN_INTEGRATION_TESTS),true)
	go test -v ./...
endif
