# Copilot Go Proxy

[🇬🇧 English](README.md) | [🇧🇷 Português](README-PT.md)

Um proxy standalone que traduz chamadas de API compatíveis com OpenAI para o GitHub Copilot. Use com OpenCode, Claude Code ou qualquer cliente compatível com OpenAI.

## Funcionalidades

- ✅ Proxy HTTP/HTTPS para GitHub Copilot
- ✅ Compatibilidade com API OpenAI (chat completions, models)
- ✅ Autenticação OAuth com GitHub (device flow)
- ✅ Atualização automática de tokens
- ✅ Suporte para streaming de respostas
- ✅ Modo daemon (systemd)
- ✅ Logging estruturado (JSON)
- ✅ Gerenciamento centralizado de configuração
- ✅ Testes unitários abrangentes

## Pré-requisitos

- Go 1.21+ (para compilar do fonte)
- GitHub account (para autenticação)
- Linux, macOS ou Windows

## Instalação

### 1. Clone o repositório

```bash
git clone https://github.com/jrperin/copilot-go-proxy.git
cd copilot-go-proxy
```

### 2. Configure variáveis de ambiente

Copie o arquivo de exemplo e configure suas credenciais:

```bash
cp .env_example .env
```

Edite `.env` e preencha com suas credenciais (a maioria dos valores já possui defaults):

```env
GITHUB_TOKEN=seu_token_aqui
GITHUB_CLIENT_ID=seu_client_id_aqui
```

⚠️ **Importante:** Nunca commit `.env` com dados reais no Git. Este arquivo está no `.gitignore`.

### 3. Compile ou execute

#### Opção A: Compilar

```bash
go build -o copilot-proxy
./copilot-proxy --help
```

#### Opção B: Executar direto

```bash
go run main.go --help
```

## Uso

### Autenticar com GitHub

Primeiro, você precisa autenticar com GitHub para obter um token:

```bash
./copilot-proxy auth
```

Siga as instruções na tela. Um código de verificação será exibido e você será direcionado para autorizar a aplicação no GitHub.

### Iniciar o proxy

Para iniciar o proxy em modo foreground:

```bash
./copilot-proxy start
```

Para iniciar como daemon (segundo plano):

```bash
./copilot-proxy start &
```

### Verificar status

```bash
./copilot-proxy status
```

### Parar o proxy

```bash
./copilot-proxy stop
```

### Reiniciar o proxy

```bash
./copilot-proxy restart
```

### Ver modelos disponíveis

```bash
./copilot-proxy models
```

### Diagnosticar problemas

```bash
./copilot-proxy diagnose
```

## Configuração com systemd

Para fazer o proxy persistir entre reinicializações do sistema:

### 1. Criar arquivo de serviço

```bash
sudo nano /etc/systemd/system/copilot-proxy.service
```

### 2. Adicionar conteúdo

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

### 3. Salvar e carregar

```bash
sudo systemctl daemon-reload
```

### 4. Ativar e iniciar

```bash
sudo systemctl enable copilot-proxy.service
sudo systemctl start copilot-proxy.service
```

### 5. Verificar status

```bash
sudo systemctl status copilot-proxy.service
```

### 6. Ver logs

```bash
sudo journalctl -u copilot-proxy.service -f
```

## Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `GITHUB_TOKEN` | Token de autenticação do GitHub | (obrigatório) |
| `GITHUB_CLIENT_ID` | ID do cliente OAuth GitHub | (obrigatório) |
| `COPILOT_API_URL` | URL da API OAuth do GitHub | `https://github.com/login/oauth/access_token` |
| `COPILOT_TOKEN_URL` | URL para obter token do Copilot | `https://api.github.com/copilot_internal/v2/token` |
| `GITHUB_DEVICE_CODE_URL` | URL para device code flow | `https://github.com/login/device/code` |

Todas as variáveis podem ser definidas em um arquivo `.env` na raiz do projeto.

## Estrutura do Projeto

```
copilot-go-proxy/
├── cmd/                    # Comandos CLI
│   ├── commands.go        # Definições dos comandos
│   └── models.go          # Gerenciamento de modelos
├── internal/
│   ├── auth/              # Autenticação OAuth e gerenciamento de tokens
│   │   ├── device.go      # Device flow OAuth
│   │   ├── token.go       # Token Manager
│   │   ├── store.go       # Armazenamento de tokens
│   │   └── token_test.go  # Testes unitários
│   ├── config/            # Configuração centralizada
│   │   └── config.go      # Carregamento de variáveis de ambiente
│   ├── logger/            # Logging estruturado
│   │   └── logger.go      # Logger JSON
│   ├── server/            # Servidor HTTP
│   │   ├── server.go      # Setup do servidor
│   │   ├── chat_completions.go # Handler de chat completions
│   │   ├── health.go      # Health check
│   │   ├── models.go      # Handler de modelos
│   │   └── messages.go    # Handler de mensagens
│   ├── copilot/           # Cliente Copilot
│   │   ├── client.go      # HTTP client
│   │   ├── headers.go     # Headers customizados
│   │   └── types.go       # Tipos de dados
│   ├── translator/        # Tradução de APIs
│   │   ├── request.go     # Tradução de requisições
│   │   ├── response.go    # Tradução de respostas
│   │   ├── streaming.go   # Streaming SSE
│   │   └── types.go       # Tipos de dados
│   └── process/           # Gerenciamento de processo
│       ├── daemon.go      # Daemon management
│       ├── pid.go         # PID file management
│       └── status.go      # Status checks
├── main.go                # Entrada da aplicação
├── go.mod                 # Dependências do Go
├── .env_example           # Exemplo de variáveis de ambiente
├── .gitignore             # Arquivos ignorados pelo Git
└── README.md              # Este arquivo
```

## Desenvolvimento

### Executar testes

```bash
go test ./...
```

Com output verboso:

```bash
go test ./... -v
```

### Formatar código

```bash
go fmt ./...
```

### Analisar código

```bash
go vet ./...
```

### Build para produção

```bash
go build -o copilot-proxy
```

## Segurança

### Proteção de Credenciais

- ✅ Token do GitHub armazenado em `~/.copilot-proxy/auth.json` (permissões 0600)
- ✅ Variáveis sensíveis carregadas de `.env` (arquivo ignorado no Git)
- ✅ `.env` e `auth.json` inclusos no `.gitignore`
- ✅ Nenhuma credencial hardcoded no código-fonte

### Logs

- ✅ Logs estruturados em JSON para facilitar análise
- ✅ Sem informações sensíveis nos logs
- ✅ Suporta diferentes níveis de log (Debug, Info, Warn, Error)

## Arquitetura de Logging

O projeto usa logging estruturado JSON via `slog` (padrão do Go 1.21+):

```bash
# Exemplo de log estruturado
{"time":"2026-05-31T22:56:42.123456Z","level":"INFO","msg":"received chat completion request","model":"gpt-4","stream":false}
{"time":"2026-05-31T22:56:42.234567Z","level":"DEBUG","msg":"token refreshed","expires_at":"2026-05-31T23:56:42Z"}
```

Logs são escritos para `stderr`, permitindo fácil redirecionamento.

## Testes

O projeto possui testes unitários abrangentes:

- **Token Manager:** Sucesso, erro de rede, erro de servidor, JSON inválido
- **Process Management:** Leitura/escrita de PID
- **Translator:** Conversão de formatos de requisição/resposta

```bash
# Rodar todos os testes
go test ./...

# Rodar com coverage
go test -cover ./...

# Rodar teste específico
go test -run TestRefreshToken_Success ./internal/auth
```

## Troubleshooting

### "not authenticated. Run: copilot-proxy auth"

Você ainda não autenticou. Execute:
```bash
./copilot-proxy auth
```

### "copilot token request failed (401)"

Seu token expirou ou é inválido. Faça autenticação novamente:
```bash
./copilot-proxy auth
```

### "No .env file found"

Isto é apenas um aviso. O proxy usará URLs padrão. Se precisar customizar, crie `.env`:
```bash
cp .env_example .env
```

### Verificar logs do systemd

```bash
sudo journalctl -u copilot-proxy.service -n 50 -f
```

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

MIT License - veja `LICENSE` para detalhes

## Referências

- [GitHub Copilot API](https://docs.github.com/en/copilot)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [Go Documentation](https://golang.org/doc)
- [systemd Documentation](https://www.freedesktop.org/wiki/Software/systemd/)

## Melhorias Recentes

Para detalhes sobre melhorias de segurança, configuração e testes, veja [IMPROVEMENTS.md](IMPROVEMENTS.md).

### Destaque das Mudanças

- ✅ **Segurança:** Credenciais isoladas em `.env`
- ✅ **Logging:** Logs estruturados em JSON
- ✅ **Configuração:** Gerenciamento centralizado
- ✅ **Testes:** 4 novos testes unitários para autenticação
- ✅ **Qualidade:** `go fmt` e `go vet` validados
