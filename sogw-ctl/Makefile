# Author: recallsong
# Email: songruiguo@qq.com

# project info
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
APP_NAME := $(shell echo ${PROJ_PATH} | sed 's/^\(.*\)[/]//')
# build info
GO_ARCH ?= $(shell go env GOARCH)
GO_OS ?= $(shell go env GOOS)
VERSION := 1.0.0
COMMIT_ID := $(shell git rev-parse HEAD 2>/dev/null)
BUILD_TIME := $(shell date "+%Y-%m-%d %H:%M:%S")
GO_VERSION := $(shell go version)
GO_SHORT_VERSION := $(shell go version | awk '{print $$3}')
VERSION_PKG := github.com/recallsong/go-utils/version
VERSION_OPS := -ldflags "\
		-X '${VERSION_PKG}.Version=${VERSION}' \
		-X '${VERSION_PKG}.BuildTime=${BUILD_TIME}' \
        -X '${VERSION_PKG}.CommitID=${COMMIT_ID}' \
        -X '${VERSION_PKG}.GoVersion=${GO_VERSION}'"

.PHONY: build-version clean

build: build-version
	go build ${VERSION_OPS} -o "${PROJ_PATH}/${APP_NAME}"

cross-build: build-version
	CGO_ENABLED=0 GOOS=${GO_OS} GOARCH=${GO_ARCH} go build ${VERSION_OPS} -o "${PROJ_PATH}/bin/${GO_OS}-${GO_ARCH}-${APP_NAME}"

build-version:
	@echo ------------ Start Build Version Details ------------
	@echo AppName: ${APP_NAME}
	@echo Arch: ${GO_ARCH}
	@echo OS: ${GO_OS}
	@echo Version: ${VERSION}
	@echo BuildTime: ${BUILD_TIME}
	@echo GoVersion: ${GO_VERSION}
	@echo CommitID: ${COMMIT_ID}
	@echo ------------ End   Build Version Details ------------

clean:
	rm -f "${PROJ_PATH}/${APP_NAME}"
	rm -rf "${PROJ_PATH}/bin"
	rm -f "${PROJ_PATH}/docker-build"
	docker rmi ${APP_NAME}:${VERSION} 2>/dev/null || :
	docker rmi ${APP_NAME}:latest 2>/dev/null || :

${APP_NAME}:
	go build ${VERSION_OPS} -o "${PROJ_PATH}/${APP_NAME}"
run-version:${APP_NAME}
	@echo "Start Run ${APP_NAME} ..."
	@"${PROJ_PATH}/${APP_NAME}" version
run-help:
	@go run main.go help
run:
	@go run main.go show --debug=true
