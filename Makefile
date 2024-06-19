BUILD_DIR=build
MODULE=github.com/infastin/wg-wish
MSGP_DIR=server/repo/db/impl/queries

deps:
	go mod tidy
.PHONY: deps

$(BUILD_DIR)/server: deps
	go build -o $@ $(MODULE)/server

build: $(BUILD_DIR)/server

ctags:
	ctags -R
.PHONY: ctags

msgp:
	go generate ./$(MSGP_DIR)
.PHONY: msgp

lint:
	golangci-lint run
.PHONY: lint
