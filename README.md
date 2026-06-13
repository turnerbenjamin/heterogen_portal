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

```terminal
go test $(go list ./internal/... | grep -E -v '/constants|/testhelpers') -cover -coverprofile=test-coverage.out
go tool cover -html=test-coverage.out
```
