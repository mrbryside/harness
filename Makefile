.PHONY: run run-log log clean build test

BUILD_DIR := build
LOG_FILE  := /tmp/harness.log
BINARY    := $(BUILD_DIR)/harness

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BINARY) .

run: build
	@$(BINARY)

run-log: build
	@> $(LOG_FILE)
	@echo "▶ TUI started — in another terminal run: make log"
	@$(BINARY) 2>>$(LOG_FILE)

log:
	@tail -f $(LOG_FILE)

test:
	@go test ./...

clean:
	@rm -rf $(BUILD_DIR) $(LOG_FILE)
