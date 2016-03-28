# dserve - Directory Serve

__dserve__ serves a specified directory via HTTP on a specified listening address

## dserve CLI Installation Options

Option 1. Download Windows or Linux executable application from dserve Github **[releases page](https://github.com/peteretelej/dserve/releases)**

Option 2. Install via `go get`. (Requires Golang)

```
go get github.com/peteretelej/dserve
```



### Usage

Run `dserve` command while in the directory to serve. Serves the current working directory on ":9011", accessible on browsers e.g via http://localhost:9011

```
cd ~/myProject
dserve
```

Or from any directory
```
dserve -d ~/myProject
``` 

- Specifying custom directory and listen address
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

