# dserve Design Document

This document covers the technical design and implementation details of dserve.

## Architecture

dserve is a single-binary HTTP file server built with Go's standard library. The architecture follows a middleware chain pattern where each feature wraps the base file server handler.

```
Request → BasicAuth → Gzip → SPA → LiveReload → FileServer/WebUI → Response
```

### Core Components

| File | Purpose |
|------|---------|
| `main.go` | Entry point, flag parsing, server setup |
| `config.go` | Configuration struct definitions |
| `ui.go` | Web UI handler and HTML embedding |
| `live.go` | Live reload via Server-Sent Events |
| `compress.go` | Gzip compression middleware |
| `spa.go` | Single-page application fallback |
| `tls.go` | TLS certificate generation |
| `upload.go` | File upload handler |
| `zip.go` | Directory zip download |

## Feature Details

### Web UI (`-webui`)

A modern directory listing interface embedded as a single HTML file.

**Implementation:**
- HTML template embedded via `//go:embed web/ui.html`
- Server injects file listing data as JSON into `window.DSERVE`
- Vanilla JS handles rendering, sorting, filtering
- CSS variables for light/dark theme support

**Features:**
- Sortable columns (name, size, date)
- Real-time search filter
- File preview (images, video, audio, PDF, text)
- Drag-and-drop upload zone (when `-upload` enabled)
- Zip download button (when `-zip` enabled)

**API:**
- `GET /` with `Accept: text/html` → HTML UI
- `GET /` with `Accept: application/json` → JSON file listing

### Live Reload (`-live`)

Browser auto-refresh when files change.

**Implementation:**
- Uses `fsnotify` for filesystem watching
- Server-Sent Events (SSE) endpoint at `/__livereload`
- JavaScript snippet injected before `</body>` in HTML responses
- 100ms debounce to batch rapid changes

**Watch Patterns:**
```bash
--live              # Watch all files (*)
--live="*.html"     # Watch only HTML files
--live="*.html,*.css,*.js"  # Multiple patterns
```

**Excluded directories:** `.git`, `node_modules`, `vendor`, dotfiles

### SPA Mode (`-spa`)

Serves a fallback file for client-side routing.

**Implementation:**
- Middleware intercepts 404 responses
- Returns fallback file (default: `index.html`) for missing paths
- Preserves actual 404 for static assets (files with extensions)

```bash
--spa              # Use index.html
--spa=app.html     # Use custom file
```

### Compression (`-compress`)

Gzip compression for text-based content.

**Compressed types:**
- `text/*` (html, css, plain, xml)
- `application/javascript`
- `application/json`
- `application/xml`
- `image/svg+xml`

**Skipped:**
- Already compressed formats (images, video, audio, fonts)
- Range requests (breaks partial content)
- Small responses (overhead not worth it)

### TLS (`-tls`)

HTTPS with automatic or custom certificates.

**Auto-generated certificates:**
- Self-signed, valid for 1 year
- Stored in `~/.config/dserve/` (Linux/macOS) or `%APPDATA%\dserve\` (Windows)
- Includes localhost and local IP addresses in SAN

**Custom certificates:**
```bash
--tls --tls-cert=server.crt --tls-key=server.key
```

### File Upload (`-upload`)

HTTP file upload via multipart form.

**Endpoint:** `POST /__upload`

**Security:**
- Filename sanitization (removes path traversal, special chars)
- Size limit via `-max-size` (default: 100MB)
- Unique filenames to prevent overwrites

**Request format:**
```
POST /__upload?path=/subdir
Content-Type: multipart/form-data

file: <binary data>
```

### Zip Download (`-zip`)

Download directories as zip archives.

**Endpoint:** `GET /__zip?path=/subdir`

**Features:**
- Streams zip directly (no temp files)
- Excludes dotfiles and hidden directories
- Preserves directory structure

### Basic Auth (`-basicauth`)

HTTP Basic Authentication.

```bash
--basicauth user:password
```

**Requirements:**
- Username: minimum 3 characters
- Password: minimum 1 character

## Configuration

### Config Struct

```go
type Config struct {
    Addr       string        // Listen address
    Timeout    time.Duration // Server timeout
    TLS        *TLSConfig    // TLS settings
    Compress   bool          // Enable gzip
    SPA        string        // SPA fallback file
    LiveReload *LiveReload   // Live reload instance
    Upload     *UploadConfig // Upload settings
    Zip        bool          // Enable zip download
    WebUI      bool          // Enable web UI
}
```

### All Flags

```
-dir string        Directory to serve (default "./")
-port int          Port to serve on (default 9011)
-local             Serve on localhost only
-timeout duration  Server timeout (default 3m0s)

-tls               Enable HTTPS
-tls-cert string   TLS certificate file
-tls-key string    TLS key file

-compress          Enable gzip compression
-spa string        SPA fallback file (default: index.html if flag present)
-live string       Live reload pattern (default: * if flag present)

-upload            Enable file uploads
-upload-dir string Upload destination directory
-max-size string   Maximum upload size (default "100MB")

-zip               Enable directory download as zip
-webui             Enable web UI for directory listing
-basicauth string  Basic auth credentials (user:pass)
```

## Internal Endpoints

| Endpoint | Purpose | Enabled By |
|----------|---------|------------|
| `/__livereload` | SSE for live reload | `-live` |
| `/__upload` | File upload | `-upload` |
| `/__zip` | Zip download | `-zip` |

## Security Considerations

1. **Path Traversal:** All file paths are sanitized and confined to the serve directory
2. **Dotfiles:** Hidden files in root are not served (configurable via Web UI)
3. **Upload Safety:** Filenames sanitized, size limits enforced
4. **HTTPS:** Auto-generated certs are self-signed (browser warning expected)

## Performance

- **Streaming:** Large files streamed, not buffered
- **Range Requests:** Supported for resumable downloads and video seeking
- **Compression:** Only applied to compressible content types
- **Debouncing:** Live reload batches rapid file changes

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/fsnotify/fsnotify` | Filesystem watching for live reload |

All other functionality uses Go standard library.

## Building

```bash
# Development
go build -o dserve .

# Release (via GoReleaser)
goreleaser release
```

## Testing

```bash
go test ./...           # Run tests
go test -cover ./...    # With coverage
go test -v ./...        # Verbose
```
