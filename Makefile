APP := zjsh
BIN_DIR := bin
INSTALL_DIR ?= $(if $(GOBIN),$(GOBIN),$(HOME)/.local/bin)

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -o $(BIN_DIR)/$(APP) ./cmd/zjsh

.PHONY: release-build
release-build:
	./scripts/build.sh

.PHONY: run
run:
	go run ./cmd/zjsh

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	gofmt -w $$(go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}{{range .TestGoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./...)

.PHONY: check
check: fmt test build

.PHONY: install
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BIN_DIR)/$(APP) $(INSTALL_DIR)/$(APP)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
