# gomultiplayer


## Getting Started

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

go get ./...
buf generate
```

## Server

```bash
go run server/*.go
```

## Client (for testing)

```bash
go run client/*.go
```

## HTML Canvas

```
open http://localhost:8080
```


