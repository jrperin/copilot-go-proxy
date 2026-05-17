# copilot-proxy

Standalone Go proxy that translates OpenAI-compatible API calls to GitHub Copilot. Single binary, zero dependencies.

Use with [OpenCode](https://opencode.ai), Claude Code, or any OpenAI-compatible client.

## Install

```bash
git clone <repo-url> && cd copilot-proxy-app
make install
```

Or build manually:

```bash
go build -o copilot-proxy .
```

## Quick Start

```bash
# 1. Authenticate (opens browser for GitHub OAuth)
copilot-proxy auth

# 2. Start the proxy (runs on localhost:4141 by default)
copilot-proxy start

# 3. Get config for your tool
copilot-proxy config
```

## Commands

| Command | Description |
|---------|-------------|
| `auth` | GitHub OAuth device flow (one-time setup) |
| `start` | Start proxy server (daemon mode) |
| `stop` | Stop proxy server |
| `restart` | Restart proxy server |
| `status` | Show running status |
| `diagnose` | Full diagnostics (auth, process, port, API health) |
| `config` | Print provider config JSON for opencode.json |

## Configure with OpenCode

Add to `~/.config/opencode/opencode.json` under `"provider"`:

```json
"copilot": {
  "npm": "@ai-sdk/openai-compatible",
  "name": "GitHub Copilot",
  "options": {
    "baseURL": "http://127.0.0.1:4141/v1",
    "apiKey": "sk-dummy"
  },
  "models": {
    "claude-sonnet-4": {
      "name": "Claude Sonnet 4 (Copilot)",
      "limit": { "context": 128000, "output": 16384 }
    },
    "gpt-4o": {
      "name": "GPT-4o (Copilot)",
      "limit": { "context": 128000, "output": 16384 }
    }
  }
}
```

Run `copilot-proxy config` to get the JSON automatically.

## How It Works

```
OpenCode / Claude Code
    |
    | POST /v1/messages  (Anthropic API format)
    v
copilot-proxy  (localhost:4141)
    |
    | Translates to OpenAI format
    | POST /chat/completions
    v
GitHub Copilot API  (api.githubcopilot.com)
```

The proxy:
1. Accepts Anthropic Messages API requests (`/v1/messages`)
2. Translates to OpenAI Chat Completions format
3. Forwards to GitHub Copilot API with proper auth headers
4. Translates responses back to Anthropic format
5. Supports both streaming (SSE) and non-streaming responses

## Configuration

| Setting | Default | Env/Flag |
|---------|---------|----------|
| Port | 4141 | `--port` / `-p` |
| Data dir | `~/.copilot-proxy/` | - |
| Token file | `~/.copilot-proxy/auth.json` | - |
| PID file | `~/.copilot-proxy/copilot-proxy.pid` | - |
| Log file | `~/.copilot-proxy/copilot-proxy.log` | - |

## Build for Multiple Platforms

```bash
make build-all
# Produces:
#   copilot-proxy-linux-amd64
#   copilot-proxy-darwin-arm64
#   copilot-proxy-windows-amd64.exe
```

## Requirements

- Go 1.21+ (for building)
- GitHub account with Copilot access
- No Node.js required
