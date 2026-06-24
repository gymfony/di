# Contributing to gymfony/di

Thank you for your interest in contributing! This guide covers everything you need to get started.

## Requirements

- Go 1.26+
- [golangci-lint](https://golangci-lint.run/welcome/install/) (latest)

## Getting started

```bash
git clone https://github.com/gymfony/di.git
cd di
go mod download
```

## Workflow

1. Fork the repository and create a branch from `main`:
   ```
   feat/<short-description>   # new feature
   fix/<short-description>    # bug fix
   docs/<short-description>   # documentation only
   refactor/<short-description>
   ```

2. Make your changes. Run the full check suite before opening a PR:
   ```bash
   bash scripts/pre-release-check.sh
   ```
   Or step by step:
   ```bash
   make fmt          # format code
   make lint         # golangci-lint
   make test         # tests with race detector
   make test-cover   # tests + coverage report
   ```

3. Ensure **all CI jobs pass** and **coverage stays at or above 85%**.

4. Open a pull request against `main` with a clear description of what and why.

## Code style

- Follow standard Go formatting — `go fmt ./...` before every commit.
- golangci-lint runs at level configured in `.golangci.yml`; all linter errors must be resolved (no `//nolint` without a comment explaining why).
- Keep functions short and focused (`funlen` limit: 80 lines / 50 statements).

## Commit messages

Use the conventional format:

```
<type>: <short summary in imperative mood>

<optional body — explain why, not what>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`.

Examples:
```
feat: add scoped lifetime support
fix: resolve circular dependency panic
docs: add usage examples to README
```

## Tests

- Every exported function must have at least one test.
- Table-driven tests are preferred for multiple input/output cases.
- Avoid mocking internal packages — test through the public API.
- Coverage must remain **≥ 85%** across the package.

## Pull request checklist

- [ ] `go fmt ./...` applied
- [ ] `go vet ./...` passes
- [ ] `golangci-lint run` passes
- [ ] All existing tests pass with `-race`
- [ ] New functionality is covered by tests
- [ ] `go.mod` / `go.sum` are tidy (`go mod tidy`)
- [ ] `README.md` updated if public API changed

## Reporting issues

Please open a GitHub issue with:
- Go version (`go version`)
- A minimal reproducer
- Expected vs actual behaviour

## License

By contributing you agree that your contributions will be licensed under the [MIT License](LICENSE).
