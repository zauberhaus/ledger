# EC core ledger

## Requirements

- [Go 1.18](https://go.dev/doc/install)
- [Docker](https://www.docker.com/)
- [docker-compose](https://docs.docker.com/compose/install/)
- [jq](https://stedolan.github.io/jq/)

## Optional

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Build

```bash
./build.sh
```

## Run the ledger with an existing immudb instance

Local immudb:

```bash
DB=test source env.sh
./core.ledger.service init // only if database doesn't exist
./core.ledger.service service
```

Immudb installed via Helm chart [core.immudb.helm](https://github.com/ec-systems/core.immudb.helm):

```bash
RELEASE={helm chart name} NS={kubernetes namespace} DB=test source env.sh
./core.ledger.service init // only if database doesn't exist
./core.ledger.service service
```

Now you should be able to access the [swagger documentation](http://localhost:8888/swagger/index.html)

## Run the ledger with docker-compose

```bash
docker-compose up
```

If you want to start a local immudb instance only, you can call:

```bash
docker-compose up immudb -d
```

## Configuration

Examples are in the folder [pkg/config/examples/conf.sample.json](https://github.com/ec-systems/core.ledger.service/tree/dev/pkg/config/examples)

## Generate files after changes

```bash
go generate ./...
```

## Run tests

The tests need a local running immudb instance. You can start one via docker-compose.

```bash
go test ./...
```

## Generate go files after protobuf changes

Requirements:

- [protoc] (http://google.github.io/proto-lens/installing-protoc.html)
- [protoc-gen-go] (https://formulae.brew.sh/formula/protoc-gen-go)

```bash
protoc --proto_path=proto --go_out=. ./proto/transaction.proto
```

## Docker build and push to GCR

```bash
docker build --platform linux/amd64 -t gcr.io/astute-synapse-332322/core-ledger-service:{version} .
docker push gcr.io/astute-synapse-332322/core-ledger-service:{version}
```

## Command line tool

```bash
./core.ledger.service --help

EC core ledger service

Usage:
core.ledger.service [command]

Available Commands:
accounts List all accounts of a holder
add Adds assets to the ledger
assets Show assets
completion Generate the autocompletion script for the specified shell
help Help about any command
history Show the history of a transaction
holders List all account holders
init Creates the database if not exists
keys Show keys of a immudb transaction
orders Show orders
remove Remove assets from the ledger
service Starts ledger web service
tx List all transactions [holder id] [asset] [account id]
version Show the version info

Flags:
--assets stringToString Supported assets (default [])
--ca string MTLs ca file name
--certificate string MTLs certificate file name
--config string Config file (default is $HOME/core.ledger.service.yaml)
-d, --database string Database name
-f, --format Format Format of the database values (default protobuf)
-h, --help help for core.ledger.service
-l, --log string Log level (error, panic, fatal, debug, info, warn) (default "info")
-m, --mtls Enable mtls
-P, --password string Database user password
--pkey string MTLs key file name
--signing-key string Path to the public key to verify signatures when presents
--statuses Statuses Supported statuses (default Unknown=-1,Created=0,CancellationFinished=998,Canceled=999,Finished=1000)
-u, --user string Database user
-v, --verbose Verbose logging
```
