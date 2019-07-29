# sentry-go

## Requirements
- Download and install Go to build and run this project, for more infomration checkout:
[Getting Started](https://golang.org/doc/install)
- There is **no need to define any Go environment variables** as the project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for packaging

## Setup & Build
Clone this repo and cd into the directory:

```
git clone https://github.com/idosun/sentry-go.git
cd sentry-go
```

To build the binary `sentry-go-demo`, create a new release, assign git commits and start the server run
```
make deploy
```

The Go HTTP Server will be available on  `http://localhost:3000`

## Demo Specs

The HTTP Server offers 3 API endpoints:
1. http://localhost:8000/handled - generates a runtime error excplicitly reported to Sentry though the SDk's captureException
2. http://localhost:8000/unhandled - generates an unhadled panic (Runtime error) reported to Sentry
3. http://localhost:8000/checkout - is used with the [Sentry REACT demo store front demo](https://github.com/sentry-demos/react)

![Sentry Go demo in action](sentry-go-demo.gif)
