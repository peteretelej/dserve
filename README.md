# dserve

Serve any directory. Static file server that is fast, zero-config, and a single binary with support for live reload, TLS, SPAs, uploads, and more.

[![CI](https://github.com/peteretelej/dserve/actions/workflows/ci.yml/badge.svg)](https://github.com/peteretelej/dserve/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/release/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/releases)
[![Go Report Card](https://goreportcard.com/badge/peteretelej/dserve)](http://goreportcard.com/report/peteretelej/dserve)
[![license](https://img.shields.io/github/license/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/blob/master/LICENSE.md)

## Install

Download the latest release for Windows, macOS, Linux, or other platforms from [Releases](https://github.com/peteretelej/dserve/releases).

Or with Go:
```bash
go install github.com/peteretelej/dserve@latest
```

## Quick Start

```bash
dserve                    # Serve current directory
dserve -dir ./public      # Serve specific directory
dserve -live              # With live reload
```

Visit http://localhost:9011

## Features

- **Zero config** - Works out of the box
- **HTTPS** - Auto-generated TLS certificates (`-tls`)
- **Live reload** - Browser refresh on file changes (`-live`)
- **SPA mode** - Fallback routing for React/Vue/etc (`-spa`)
- **File uploads** - Drag & drop via web UI (`-upload`)
- **Directory download** - Download folders as zip (`-zip`)
- **Compression** - Gzip for text content (`-compress`)
- **Basic auth** - Password protection (`-basicauth`)
- **Web UI** - Modern directory listing with dark mode (`-webui`)

## Examples

```bash
# Development with live reload
dserve -live -webui

# Single-page application
dserve -spa -live

# Share files on local network
dserve -webui -upload -zip

# Secure with HTTPS and auth
dserve -tls -basicauth admin:secret123
```

## All Flags

```
dserve -help
  -basicauth string
    	basic auth credentials (user:pass)
  -cert string
    	TLS certificate file
  -compress
    	enable gzip compression
  -dir string
    	directory to serve (default "./")
  -key string
    	TLS key file
  -live string
    	enable live reload with watch pattern (default: * if flag present)
  -local
    	serve on localhost only
  -max-size string
    	maximum upload size (default "100MB")
  -port int
    	port to serve on (default 9011)
  -spa string
    	enable SPA mode with fallback file (default: index.html if flag present)
  -timeout duration
    	server timeout (default 3m0s)
  -tls
    	enable HTTPS
  -upload
    	enable file uploads
  -upload-dir string
    	upload destination directory
  -webui
    	enable web UI for directory listing
  -zip
    	enable directory download as zip
```

## Documentation

See [docs/design.md](docs/design.md) for technical details.

## Requirements

- Go 1.21+ (for building from source)
- Windows 10+ / macOS / Linux

> Windows 7/8 users: use [v2.2.4](https://github.com/peteretelej/dserve/releases/tag/v2.2.4)

## License

MIT
