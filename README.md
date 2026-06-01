# Copilot Go Proxy

[🇬🇧 English](README.md) | [🇧🇷 Português](README-PT.md)

A standalone proxy that translates OpenAI-compatible API calls to GitHub Copilot. Use with OpenCode, Claude Code, or any OpenAI-compatible client.

## Features

- ✅ HTTP/HTTPS proxy for GitHub Copilot
- ✅ OpenAI API compatibility (chat completions, models)
- ✅ GitHub OAuth authentication (device flow)
- ✅ Automatic token refresh
- ✅ Response streaming support
- ✅ Daemon mode (systemd)
- ✅ Structured JSON logging
- ✅ Centralized configuration management
- ✅ Comprehensive unit tests

## Prerequisites

- Go 1.21+ (to compile from source)
- GitHub account (for authentication)
- Linux, macOS, or Windows

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/jrperin/copilot-go-proxy.git
cd copilot-go-proxy
```

### 2. Configure environment variables

Copy the example file and configure your credentials:

```bash
cp .env_example .env
```

Edit `.env` and fill in your credentials (most values already have defaults):

```env
GITHUB_TOKEN=your_token_here
GITHUB_CLIENT_ID=your_client_id_here
```

⚠️ **Important:** Never commit `.env` with real data to Git. This file is in `.gitignore`.

### 3. Compile or run

#### Option A: Compile

```bash
go build -o copilot-proxy
./copilot-proxy --help
```

#### Option B: Run directly

```bash
go run main.go --help
```

## Usage

### Authenticate with GitHub

First, you need to authenticate with GitHub to get a token:

```bash
./copilot-proxy auth
```

Follow the instructions on screen. A verification code will be displayed and you will be directed to authorize the application on GitHub.

### Start the proxy

To start the proxy in foreground mode:

```bash
./copilot-proxy start
```

To start as a daemon (background):

```bash
./copilot-proxy start &
```

### Check status

```bash
./copilot-proxy status
```

### Stop the proxy

```bash
./copilot-proxy stop
```

### Restart the proxy

```bash
./copilot-proxy restart
```

### View available models

```bash
./copilot-proxy models
```

### Diagnose issues

```bash
./copilot-proxy diagnose
```

## Systemd Configuration

To make the proxy persist across system reboots:

### 1. Create service file

```bash
sudo nano /etc/systemd/system/copilot-proxy.service
```

### 2. Add content

```ini
[Unit]
Description=GitHub Copilot Proxy
After=network.target

[Service]
Type=simple
User=your-username
Environment="PATH=/home/your-username/.local/bin:/usr/local/bin:/usr/bin:/bin"
ExecStart=/home/your-username/.local/bin/copilot-proxy start --foreground
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 3. Save and reload

```bash
sudo systemctl daemon-reload
```

### 4. Enable and start

```bash
sudo systemctl enable copilot-proxy.service
sudo systemctl start copilot-proxy.service
```

### 5. Check status

```bash
sudo systemctl status copilot-proxy.service
```

### 6. View logs

```bash
sudo journalctl -u copilot-proxy.service -f
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GITHUB_TOKEN` | GitHub authentication token | (required) |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | (required) |
| `COPILOT_API_URL` | GitHub OAuth API URL | `https://github.com/login/oauth/access_token` |
| `COPILOT_TOKEN_URL` | Copilot token endpoint URL | `https://api.github.com/copilot_internal/v2/token` |
| `GITHUB_DEVICE_CODE_URL` | Device code flow URL | `https://github.com/login/device/code` |

All variables can be defined in a `.env` file at the project root.

## Project Structure

```
copilot-go-proxy/
├── cmd/                    # CLI commands
│   ├── commands.go        # Command definitions
│   └── models.go          # Model management
├── internal/
│   ├── auth/              # OAuth authentication and token management
│   │   ├── device.go      # Device flow OAuth
│   │   ├── token.go       # Token Manager
│   │   ├── store.go       # Token storage
│   │   └── token_test.go  # Unit tests
│   ├── config/            # Centralized configuration
│   │   └── config.go      # Environment variable loading
│   ├── logger/            # Structured logging
│   │   └── logger.go      # JSON logger
│   ├── server/            # HTTP server
│   │   ├── server.go      # Server setup
│   │   ├── chat_completions.go # Chat completions handler
│   │   ├── health.go      # Health check
│   │   ├── models.go      # Models handler
│   │   └── messages.go    # Messages handler
│   ├── copilot/           # Copilot client
│   │   ├── client.go      # HTTP client
│   │   ├── headers.go     # Custom headers
│   │   └── types.go       # Data types
│   ├── translator/        # API translation
│   │   ├── request.go     # Request translation
│   │   ├── response.go    # Response translation
│   │   ├── streaming.go   # SSE streaming
│   │   └── types.go       # Data types
│   └── process/           # Process management
│       ├── daemon.go      # Daemon management
│       ├── pid.go         # PID file management
│       └── status.go      # Status checks
├── main.go                # Application entry point
├── go.mod                 # Go dependencies
├── .env_example           # Example environment variables
├── .gitignore             # Git ignore rules
└── README.md              # This file
```

## Development

### Run tests

```bash
go test ./...
```

With verbose output:

```bash
go test ./... -v
```

### Format code

```bash
go fmt ./...
```

### Analyze code

```bash
go vet ./...
```

### Production build

```bash
go build -o copilot-proxy
```

## Security

### Credential Protection

- ✅ GitHub token stored in `~/.copilot-proxy/auth.json` (permissions 0600)
- ✅ Sensitive variables loaded from `.env` (file ignored in Git)
- ✅ `.env` and `auth.json` included in `.gitignore`
- ✅ No credentials hardcoded in source code

### Logs

- ✅ Structured logs in JSON for easy analysis
- ✅ No sensitive information in logs
- ✅ Supports different log levels (Debug, Info, Warn, Error)

## Logging Architecture

The project uses structured JSON logging via `slog` (Go 1.21+ standard):

```bash
# Example structured log
{"time":"2026-05-31T22:56:42.123456Z","level":"INFO","msg":"received chat completion request","model":"gpt-4","stream":false}
{"time":"2026-05-31T22:56:42.234567Z","level":"DEBUG","msg":"token refreshed","expires_at":"2026-05-31T23:56:42Z"}
```

Logs are written to `stderr`, allowing easy redirection.

## Tests

The project has comprehensive unit tests:

- **Token Manager:** Success, network error, server error, invalid JSON
- **Process Management:** PID read/write operations
- **Translator:** Request/response format conversion

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestRefreshToken_Success ./internal/auth
```

## Troubleshooting

### "not authenticated. Run: copilot-proxy auth"

You haven't authenticated yet. Execute:
```bash
./copilot-proxy auth
```

### "copilot token request failed (401)"

Your token has expired or is invalid. Authenticate again:
```bash
./copilot-proxy auth
```

### "No .env file found"

This is just a warning. The proxy will use default URLs. If you need to customize, create `.env`:
```bash
cp .env_example .env
```

### Check systemd logs

```bash
sudo journalctl -u copilot-proxy.service -n 50 -f
```

## Contributing

1. Fork the project
2. Create a branch for your feature (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

MIT License - see `LICENSE` file for details

## References

- [GitHub Copilot API](https://docs.github.com/en/copilot)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Go Documentation](https://golang.org/doc)
- [systemd Documentation](https://www.freedesktop.org/wiki/Software/systemd/)

## Recent Improvements

For details about security, configuration, and testing improvements, see [IMPROVEMENTS.md](IMPROVEMENTS.md).

### Highlights of Recent Changes

- ✅ **Security:** Credentials isolated in `.env`
- ✅ **Logging:** Structured JSON logging
- ✅ **Configuration:** Centralized management
- ✅ **Tests:** 4 new unit tests for authentication
- ✅ **Quality:** `go fmt` and `go vet` validated
