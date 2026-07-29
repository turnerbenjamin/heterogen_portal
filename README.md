# Heterogen (Go + HTMX)

A seed management app to track the lineage of heterogeneous seed crops. This
project is a rewrite of an existing javascript application written with Node and
React. The aim is to add a few new features and improve maintainability by using
strongly-typed languages.

Quick overview

- **Languages**: Go + Typescript
- **Frameworks**: HTMX
- **Cloud**: Azure AD, Azure SQL, Azure Container Apps

## Unit Tests

To run unit tests and view test coverage for the project:

```bash
go test ./internal/*** -cover -coverprofile=test-coverage.out && \
go tool cover -html=test-coverage.out
```

### Mocks

Mocks are generated using [Mockery](https://github.com/vektra/mockery), the
config can be found in .mockery.yml. To generate mocks install mockery and run:

```bash
mockery
```

## Model Builder

There is a model builder which reads the database schema and generates model and
metadata structs.

```bash
go run ./cmd/model-builder/main.go --dsn=<YOUR_DSN>
```
