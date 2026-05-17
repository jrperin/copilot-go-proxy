NAME=copilot-proxy
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w"
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: build install uninstall clean test status diagnose start stop restart auth config

build:
	go build $(LDFLAGS) -o $(NAME) .

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(NAME) $(INSTALL_DIR)/$(NAME)
	@echo "Installed to $(INSTALL_DIR)/$(NAME)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH:"
	@echo "  export PATH=\"$(INSTALL_DIR):\$$PATH\""
	@echo ""
	@echo "Next steps:"
	@echo "  1. $(NAME) auth     # GitHub OAuth login"
	@echo "  2. $(NAME) start    # Start the proxy"
	@echo "  3. $(NAME) config   # Get config for opencode.json"

uninstall:
	@$(MAKE) stop 2>/dev/null || true
	rm -f $(INSTALL_DIR)/$(NAME)
	@echo "Uninstalled $(INSTALL_DIR)/$(NAME)"

clean:
	rm -f $(NAME)
	go clean

test:
	go vet ./...
	go test ./... -v

start:
	./$(NAME) start

stop:
	./$(NAME) stop

restart:
	./$(NAME) restart

status:
	./$(NAME) status

diagnose:
	./$(NAME) diagnose

auth:
	./$(NAME) auth

config:
	./$(NAME) config

build-all:
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o $(NAME)-linux-amd64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o $(NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(NAME)-windows-amd64.exe .
	@echo "Binaries built:"
	@ls -lh $(NAME)-*
