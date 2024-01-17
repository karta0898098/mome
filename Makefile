TARGET := mome
SRC = $(shell find ./cmd -name ${TARGET})
COMMIT_ID = $(shell git rev-parse --short HEAD)
VERSION := $(if $(VERSION),$(VERSION),${COMMIT_ID})
BUILD_FLAGS =-ldflags "-X main.Version=${VERSION} -X "main.CommitID='${COMMIT_ID}'


.PHONY: hello
hello:
	@echo TARGET=${TARGET}
	@echo SRC=${SRC}
	@echo BUILD_FLAGS=${BUILD_FLAGS}

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: linux
linux: ${TARGET}.linux

.PHONY: linux.arm
linux.arm: ${TARGET}.linux.arm

.PHONY: zip
zip:
	zip ${TARGET} ${TARGET}

.PHONY: build
build:
	go build ${BUILD_FLAGS} -o ${TARGET} ${SRC}

${TARGET}.linux: $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${TARGET} ${SRC}

${TARGET}.linux.arm: $(SRC)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${TARGET} ${SRC}