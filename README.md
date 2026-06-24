# gymfony/di

Dependency Injection Container for the [Gymfony](https://github.com/gymfony) framework. Go · zero dependencies.

[![CI](https://github.com/gymfony/di/actions/workflows/ci.yml/badge.svg)](https://github.com/gymfony/di/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/gymfony/di/branch/main/graph/badge.svg)](https://codecov.io/gh/gymfony/di)
[![Go Reference](https://pkg.go.dev/badge/github.com/gymfony/di.svg)](https://pkg.go.dev/github.com/gymfony/di)

## Installation

```bash
go get github.com/gymfony/di
```

Requires Go 1.26+.

## Quick start

```go
package main

import (
    "fmt"

    "github.com/gymfony/di"
)

func main() {
    c := di.New()

    // Register a value
    di.Register(c, func() *MyService {
        return &MyService{}
    })

    // Resolve
    svc := di.Resolve[*MyService](c)
    fmt.Println(svc)
}
```

## Development

```bash
# Run tests
make test

# Run tests with coverage report (target ≥85%)
make test-cover

# Lint
make lint

# Format
make fmt

# Full pre-release check
bash scripts/pre-release-check.sh
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
