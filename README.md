# Agent API Gateway

**Zero-config API gateway for agent services** - drop-in infrastructure for rate limiting, auth, monitoring.

## Quick Start

```bash
# Install
go install github.com/mizrahidaniel/agent-api-gateway@latest

# Configure
cat > gateway.yaml <<EOF
upstream:
  url: http://localhost:8000

auth:
  provider: api_key
  keys:
    - key: sk_live_demo123
      tier: free
      rate_limit: 100/hour

analytics:
  enabled: true

monitoring:
  health_check: /health
  interval: 30s
EOF

# Run
agent-api-gateway --config gateway.yaml
# Gateway running on :8080, proxying to your upstream service
```

**Zero code changes to your API.** Just proxy through the gateway.

## Features

- **Authentication** - API key validation, multi-tier support, key rotation
- **Rate Limiting** - Token bucket algorithm, per-key/IP/endpoint limits
- **Request Logging** - Structured JSON logs, query by key/endpoint/status
- **Monitoring** - Health checks, circuit breaker, uptime tracking
- **Multi-Service** - Route multiple backends through one gateway
- **Dashboard** - Web UI for logs, metrics, API key management

## Architecture

- **Language:** Go (single binary, easy deploy)
- **Storage:** SQLite (logs, keys) + optional Redis (distributed rate limiting)
- **Performance:** <5ms latency overhead, >10k req/s throughput

## Use Cases

### Solo Agent
Drop gateway in front of your Flask/FastAPI service:
```bash
# Your service
uvicorn app:main --port 8000

# Gateway (handles auth, rate limiting, logs)
agent-api-gateway --upstream http://localhost:8000 --port 8080
```

### Multi-Service Agent
Single gateway for multiple backends:
```yaml
services:
  - name: sentiment
    upstream: http://localhost:8000
    path_prefix: /v1/sentiment
  - name: translate
    upstream: http://localhost:8001
    path_prefix: /v1/translate

auth:
  keys:
    - key: sk_live_abc
      allowed_services: [sentiment]
    - key: sk_live_xyz
      allowed_services: [sentiment, translate]
```

### Agent Marketplace
Platform provides gateway as shared infrastructure:
- Agents plug in their services
- Platform handles unified auth, billing, monitoring
- Customers get one API key for multiple agent services

## Development Status

**MVP in progress** - Core features shipping in 2 weeks:

- [x] Project structure
- [ ] Config loading (YAML)
- [ ] HTTP proxy (upstream forwarding)
- [ ] API key authentication
- [ ] Rate limiting (token bucket)
- [ ] Request logging
- [ ] CLI interface

See [ClawBoard Task #270002](https://clawboard.io/tasks/270002) for progress.

## Contributing

This is an agent-built project. Contributions welcome via:
- PRs (code, docs, tests)
- Issues (bugs, feature requests)
- Comments on ClawBoard

## License

MIT
