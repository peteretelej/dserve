# dserve

[![CI](https://github.com/peteretelej/dserve/actions/workflows/ci.yml/badge.svg)](https://github.com/peteretelej/dserve/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/peteretelej/dserve)](http://goreportcard.com/report/peteretelej/dserve)

A fast, zero-config HTTP file server for local development.

## Install

```bash
go install github.com/peteretelej/dserve@latest
```

Or download from [Releases](https://github.com/peteretelej/dserve/releases).

## Quick Start

```bash
# Serve current directory
dserve

# Serve a specific directory
dserve -dir ./public

# With live reload for development
dserve -live

# Enable all features
dserve -webui -live -upload -zip -compress
```

Visit http://localhost:9011

## Features

| Flag | Description |
|------|-------------|
| `-dir` | Directory to serve (default: current) |
| `-port` | Port number (default: 9011) |
| `-local` | Bind to localhost only |
| `-webui` | Modern directory listing UI |
| `-live` | Auto-reload browser on file changes |
| `-spa` | Single-page app mode (fallback to index.html) |
| `-upload` | Enable file uploads |
| `-zip` | Enable directory download as zip |
| `-compress` | Gzip compression |
| `-tls` | HTTPS with auto-generated certificate |
| `-basicauth` | HTTP basic auth (user:pass) |

## Examples

**Development server with live reload:**
```bash
dserve -live -webui
```

**SPA development (React, Vue, etc):**
```bash
dserve -spa -live
```

**Share files on local network:**
```bash
dserve -webui -upload -zip
```

**Secure with HTTPS and auth:**
```bash
dserve -tls -basicauth admin:secret123
```

## Documentation

See [docs/design.md](docs/design.md) for technical details.

## Requirements

- Go 1.21+ (for building)
- Windows 10+ / macOS / Linux

## License

MIT - See [LICENSE](LICENSE.md)
