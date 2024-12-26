# Simple Web Server

Compile HTML and MD template files together into a static html website.

This module uses [gofiber](https://github.com/gofiber/fiber) and [staticweb](https://github.com/tkdeng/staticweb/) to create a quick and easy setup for a simple web server.

For extra performance, consider adding the `pages.dist` directory to a cdn (like cloudflare pages) to serve static html pages.

## Installation

```shell
# install the go module
go get github.com/tkdeng/simplewebserver

# install dependencies
make
```

## Dependencies

### Debian/Ubuntu (Linux)

```shell script
  sudo apt install libpcre3-dev
```

### Fedora (Linux)

```shell script
  sudo dnf install pcre-devel
```

### Arch (Linux)

```shell script
  sudo yum install pcre-dev
```

## Usage

```go

import (
  server "github.com/tkdeng/simplewebserver"
)

func main(){
  // create new server
  app, err := server.New("./app")

  //note: page.dist files will automatically be statically rendered,
  // and take priority over gofiber methods

  // do normal gofiber stuff (optional)
  app.Get("/api", func(c fiber.Ctx) error {
    return c.SendString("Hello, API!")
  })

  //note: page.dist files will automatically be statically rendered,
  // and take priority over gofiber methods
  app.Get("/", func(c fiber.Ctx) error {
    // this will be ignored if index.html exists
    return c.SendString("Hello, World!")
  })

  // listen with openssl (default port: [http: 8080, ssl: 8443])
  err = app.Listen()
}
```

## Inside App Directory

### config.yml

```yaml
title: "Web Server"
app_title: "WebServer"
desc: "A Web Server."

public_uri: "/public/"

port_http: 8080
port_ssl: 8443

origins: [
  "localhost",
  "example.com",
]

proxies: [
  "127.0.0.1",
  "192.168.0.1",
]

DebugMode: no
```
