# dserve - Directory Serve

[![Build Status](https://travis-ci.org/peteretelej/dserve.svg?branch=master)](https://travis-ci.org/peteretelej/dserve)
[![GitHub release](https://img.shields.io/github/release/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/releases)
[![Go Report Card](https://goreportcard.com/badge/peteretelej/dserve)](http://goreportcard.com/report/peteretelej/dserve)
[![license](https://img.shields.io/github/license/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/blob/master/LICENSE.md)

__dserve__ serves a specified static directory on the web 

## dserve Installation 

#### Option 1 (Fast & Easy)
Download Windows, Linux or Mac 32bit or 64bit executable application from the releases:

   - **[Download dserve](https://github.com/peteretelej/dserve/releases)**

#### Option 2
Install via `go get`. (Requires Golang)

```
go get github.com/peteretelej/dserve
```

### Usage
```
dserve serves a static directory over http

Usage:
        dserve
        dserve [flags].. [directory]

Examples:
        dserve                  Serves the current directory over http at :9011
        dserve -local           Serves the current directory on localhost:9011
        dserve -dir ~/dir       Serves the directory ~/dir over http 
        dserve -basic "guest:Pass1234"
		 Serves the current directory with basicauth (only use this over  https)

Flags:
  -dir string
        the directory to serve, defaults to current directory (default "./")
  -local bool
	whether to only serve on localhost
  -port int
        the port to serve at, defaults 9011 (default 9011)
  -timeout duration
        http server read timeout, write timeout will be double this (default 3m0s)
  -basicauth string
	enable HTTP basic authentication, arguments should be USERNAME:PASSWORD 
	example: dserve -basicauth "admin:passw0rd"
```


