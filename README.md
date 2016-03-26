# dserve - Directory Serve

__dserve__ serves a specified directory via HTTP on a specified listening address

## dserve CLI Installation Options

- Install via `go get`

```
go get github.com/peteretelej/dserve
```

- Download executable binary from Github releases page


### Usage

Run `dserve` command while in the directory to serve. Serves the current working directory on ":9011", accessible on browsers e.g via http://localhost:9011

```
go get bitbucket.org/etelej/dserve
cd ~/myProject
dserve
```

Specifying custom directory and listen address
```
dserve -d /home/chief/mystaticwebsite -l 8011
# Note: serving on port 80 requires root
```

- `dserve --help` for cli usage help
- `-l`, `--listen-addr` - custom listen address
- `-d`, `--directory` - custom directory to serve

## dserve go package

Get the package: `package github.com/peteretelej/dserve/dserve`

```
go get github.com/peteretelej/dserve/dserve
```

Import into your code
```
import "github.com/peteretelej/dserve/dserve"
```

Launch the server with `dserve.Serve(directory,listenAddress)` where _directory_ is a string path to the folder to serve, and _listenAddress_ is the address to listen on.

Example
```
package main

import "github.com/peteretelej/dserve/dserve"

func main() {
	// Serving contents of current folder on port 80
	dserve.Serve(".",":80")
}
```

