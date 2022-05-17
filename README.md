# net-dpservice-go
[![Pull Request Code test](https://github.com/onmetal/net-dpservice-go/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/onmetal/partitionlet/actions/workflows/test.yml)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://makeapullrequest.com)
[![GitHub License](https://img.shields.io/static/v1?label=License&message=Apache-2.0&color=blue&style=flat-square)](LICENSE)

Golang bindings for the [net-dpservice](https://github.com/onmetal/net-dpservice).

## Development

To regenerate the golang bindings run

```shell
make clean generate
```

## Usage

```go
package main

import (
    "context"
    dpdkproto "github.com/onmetal/net-dpservice-go/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    ctx := context.Background()
    conn, err := grpc.DialContext(ctx, "127.0.0.1", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
    if err != nil {
    panic("aaaahh")
    }
    client := dpdkproto.NewDPDKonmetalClient(conn)
    ...
}
```

## Contributing

We'd love to get feedback from you. Please report bugs, suggestions or post questions by opening a GitHub issue.

## License

[Apache-2.0](LICENSE)