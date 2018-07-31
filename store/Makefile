# project info
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
GOPATH_FIRST := $(shell echo $(subst :, ,$(GOPATH)) | awk '{print $$1}')
GOGOPROTO_PATH="${GOPATH_FIRST}/src/github.com/gogo/protobuf"

.PHONY: clean

build: clean
	protoc -I=.:"${GOPATH_FIRST}/src":"${GOGOPROTO_PATH}":"${GOGOPROTO_PATH}/protobuf" --gofast_out=. meta/meta.proto

clean:
	cd "${PROJ_PATH}/meta/" && rm -f *.pb.go