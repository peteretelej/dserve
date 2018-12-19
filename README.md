# dserve - Directory Serve

[![Build Status](https://travis-ci.org/peteretelej/dserve.svg?branch=master)](https://travis-ci.org/peteretelej/dserve)
[![GitHub release](https://img.shields.io/github/release/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/releases)
[![Go Report Card](https://goreportcard.com/badge/peteretelej/dserve)](http://goreportcard.com/report/peteretelej/dserve)
[![license](https://img.shields.io/github/license/peteretelej/dserve.svg)](https://github.com/peteretelej/dserve/blob/master/LICENSE.md)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fpeteretelej%2Fdserve?ref=badge_shield)

__dserve__ serve a directory over HTTP

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