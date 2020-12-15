# sentry-go

## Requirements

- Download and install Go to build and run this project, for more infomration checkout: [Getting Started](https://golang.org/doc/install)
- There is **no need to define any Go environment variables** as the project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for packaging

## Setup & Build

- Clone this repo and cd into the directory
- Run `make deploy` to build the binary `sentry-go-demo`, create a new release, assign git commits and start the server run
- The Go HTTP Server will be available on `http://localhost:3002`

## Demo Specs

The demo initializes Sentry SDK through the Sentry Client and then uses the [net/http](https://docs.sentry.io/platforms/go/http/) integration to attach a Sentry handler for all endpoint requests.
The HTTP Server offers 4 API endpoints:
1. http://localhost:3000/handled - generates a runtime error excplicitly reported to Sentry though the SDk's captureException. level:error
2. http://localhost:3000/unhandled - generates an unhadled panic (Runtime error) reported to Sentry. level:fatal
3. http://localhost:3000/checkout - is used with the [Sentry REACT demo store front demo](https://github.com/sentry-demos/react)
4. http://localhost:3000/success - is for generating a Transaction

![Sentry Go demo in action](sentry-go-demo.gif)
