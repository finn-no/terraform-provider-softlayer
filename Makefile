all: build test

build:
	go get && go build -v -o "terraform-provider-softlayer-${GIMME_OS}-${GIMME_ARCH}${EXT}"

test:
ifeq ($GIMME_OS,linux)
	go test -v ./...
endif
