package main

import "time"

type Config struct {
	Addr       string
	Timeout    time.Duration
	TLS        *TLSConfig
	Compress   bool
	SPA        string // empty = disabled, otherwise fallback file
	LiveReload *LiveReload
	Upload     *UploadConfig
	Zip        bool
	WebUI      bool
}

type TLSConfig struct {
	Cert string
	Key  string
}

type UploadConfig struct {
	Dir      string
	MaxBytes int64
}
