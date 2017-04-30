[![Build Status](https://travis-ci.org/peteretelej/dserve.svg?branch=master)](https://travis-ci.org/peteretelej/dserve)

# dserve - Directory Serve

[![Join the chat at https://gitter.im/dserve-app/Lobby](https://badges.gitter.im/dserve-app/Lobby.svg)](https://gitter.im/dserve-app/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

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


serve a directory
```
dserve --dir ~/myProject
``` 

serve current directory on a specific address
```
dserve --port 8011
```

serve current directory on localhost
```
dserve --local
```

serve current directory as well as a basic_auth secured directory secure/static
```
dserve --secure
```


- Specifying custom directory and listen address, on localhost
```
dserve --dir /home/chief/mystaticwebsite --port 8011 --local
# Note: serving on port 80 requires admin rights
```

`dserve --help` for cli usage help

- `--port` - custom port to listen on, default is 9011
- `--dir` - custom directory to serve, default is the directory dserve is run from
- `--local` - only serve on localhost
- `--secure` - serve a HTTP basic_auth secured directory at secure/static


## dserve go package
Get: `go get github.com/peteretelej/dserve`

Usage Example: 
```
package main

import "github.com/peteretelej/dserve"

func main() {
	// Serving contents of current folder on port 8011
	dserve.Serve(".",":8011")
}
```

## Secure directory
The secure directory (served at secure if the `--secure` flag is used) uses __http basic authentication__. Files are served from the `secure/static` directory (relative to current directory `--dir`) and server on `/secure/`

Configuration:
A sample configuration file is the secure folder (`securepass.json.sample`). Copy the sample file and rename to `securepass.json` and edit the credentials as required.
	- secure/securepass.json.sample - a sample username and password 
	- secure/securepass.json - your username and password ( create this file , both username and password in plain text)

Changing of the configuration file (e.g password) does not require restart to pick ne crendentials.

