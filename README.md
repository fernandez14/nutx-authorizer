## Nutx
Nu TX is Go program which authorizes transactions for a specific account following a set
of predefined rules.

It is a proof of concept, and for now it only reads from stdin and writes to stdout.

## Overview
NuTX is a simple in-memory transaction authorizer. It was written with testability and extendability in mind.

## Requirements
- Go 1.16

## Structure of Go packages
Project structure (mostly) follows [Standard Go Project Layout](https://github.com/golang-standards/project-layout).
- `cmd/*` - main application entry point
- `docs/*` - additional documentation
- `internal/authorizer` - authorizer system

## Features
- Account creation.
- Transaction authorization for the account.
- Extendable business logic rules.

# Build and run
```
go build -o authorize cmd/nutx/*

./authorize < operations
```


