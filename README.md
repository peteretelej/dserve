# dserve - Directory Serve

[![CI](https://github.com/peteretelej/dserve/actions/workflows/ci.yml/badge.svg)](https://github.com/peteretelej/dserve/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/release/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/releases)
[![Go Report Card](https://goreportcard.com/badge/peteretelej/dserve)](http://goreportcard.com/report/peteretelej/dserve)
[![license](https://img.shields.io/github/license/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/blob/master/LICENSE.md)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve?ref=badge_shield)

__dserve__ serve a directory over HTTP

## Installation

### Option 1: Go Install (Recommended)

Requires Go 1.24+

```bash
go install github.com/peteretelej/dserve@latest
```

### Option 2: Download Binary

Download from [Releases](https://github.com/peteretelej/dserve/releases)

> **Note:** Requires Windows 10 or later. For Windows 7/8, use [v2.2.4](https://github.com/peteretelej/dserve/releases/tag/v2.2.4)

### Usage
Enter the directory you'd like to serve and run
```
dserve
```

That's it. This launches a webserver on port 9011 serving the directory. Visit [http://localhost:9011](http://localhost:9011) to access the site.

Speficy a directory in another location
```
dserve -dir /var/www/html
```

Serve on a different port
```
dserve -port 8080
```

Enable HTTP basic authentication
```
dserve -basicauth user1:pass123
```

Restrict server to localhost
```
dserve -local
```

You can chain the arguments
```
dserve -dir ~/mysite -port 80 -basicauth user:pass12345
```



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve?ref=badge_large)