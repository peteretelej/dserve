[![Build Status](https://travis-ci.org/peteretelej/dserve.svg?branch=master)](https://travis-ci.org/peteretelej/dserve)

# dserve - Directory Serve

__dserve__ serves a specified static directory on the web 

## dserve Installation Installation Options

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
        dserve -secure          Serves the current directory with basicauth using sample .basicauth.json
        dserve -secure -basicauth myauth.json
                                Serves the current directory with basicauth using config file myauth.json

Flags:
  -basicauth string
        file to be used for basicauth json config (default ".basicauth.json")
  -dir string
        the directory to serve, defaults to current directory (default "./")
  -local
        whether to serve on all address or on localhost, default all addresses
  -port int
        the port to serve at, defaults 9011 (default 9011)
  -secure
        whether to create a basic_auth secured secure/ directory, default false
  -timeout duration
        http server read timeout, write timeout will be double this (default 3m0s)
```


