# Agent API Gateway

Zero-config API gateway for AI services. Single binary, simple YAML config.

## Features

- **Reverse Proxy:** Route `/service-name/*` to backend services
- **Authentication:** Bearer tokens or API keys
- **Rate Limiting:** Per-service, per-IP limits (requests/minute)
- **Graceful Shutdown:** Clean shutdown on SIGTERM/SIGINT

## Quick Start

```bash
# Install
go install github.com/mizrahidaniel/agent-api-gateway@latest

# Create config
cp gateway.example.yaml gateway.yaml

# Run
agent-api-gateway gateway.yaml
```

## Configuration

```yaml
port: 8080

services:
  # Public service, no auth
  public-api:
    target: "http://localhost:3000"

  # Authenticated + rate limited
  ai-service:
    target: "http://localhost:4000"
    auth:
      type: bearer  # or "apikey"
      tokens:
        - "secret-token-123"
    rate_limit:
      requests_per_minute: 60
```

## Usage

```bash
# Route to ai-service
curl -H "Authorization: Bearer secret-token-123" \
  http://localhost:8080/ai-service/v1/generate

# Route to public-api (no auth)
curl http://localhost:8080/public-api/healthcheck
```

## Path Rewriting

Requests to `/service-name/path` are proxied to `target + /path`.

Example:
- Request: `GET /ai-service/v1/models`
- Proxied to: `GET http://localhost:4000/v1/models`

## Auth Types

### Bearer Token
```yaml
auth:
  type: bearer
  tokens:
    - "token-1"
    - "token-2"
```

Send: `Authorization: Bearer token-1`

### API Key
```yaml
auth:
  type: apikey
  tokens:
    - "key-abc"
```

Send: `X-API-Key: key-abc`

## Rate Limiting

Per-service, per-client-IP. Returns `429 Too Many Requests` when exceeded.

```yaml
rate_limit:
  requests_per_minute: 100
```

Response headers:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `Retry-After`

## Why This?

Every agent service rebuilds the same infrastructure. This gives you:
- ✅ Auth
- ✅ Rate limiting
- ✅ Multi-service routing
- ✅ Single binary, zero dependencies

Ship services faster. Let the gateway handle the boring stuff.

## License

MIT
