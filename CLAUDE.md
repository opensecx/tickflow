# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TickFlow Go SDK (`github.com/opensecx/tickflow`) — a Go client library wrapping the [TickFlow financial data API](https://api.tickflow.org) (stocks, ETFs, indices, bonds, funds, options across US/CN/HK markets). All API endpoints are authenticated via the `x-api-key` header.

- Module: `github.com/opensecx/tickflow`, Go 1.26.5
- HTTP client: [`carlmjohnson/requests`](https://github.com/carlmjohnson/requests) (fluent builder for requests + JSON decode)
- Testing: `stretchr/testify` (assert + require)
- Lint: `golangci-lint` v2.12

## Commands

```bash
# Run all tests
go test -v ./...

# Run a single test by name
go test -v -run TestGetQuote ./...

# Run tests for one file's functionality (match the test function prefix)
go test -v -run 'TestGetKline|TestBatchGetKline|TestGetExFactor' ./...

# Build (library — verifies compilation)
go build ./...

# Lint (requires golangci-lint installed)
golangci-lint run
```

## Architecture

### Single-package, domain-split layout

All source lives in one package `tickflow` at the repo root. Files are organized by API domain:

| File | Domain | Key types / methods |
|------|--------|---------------------|
| `tickflow.go` | Client core | `TickFlow` struct, `NewTickFlow(key)`, shared errors, `validExchanges` |
| `model.go` | Shared models | `Instrument`, `InstrumentType`, `InstrumentExt` |
| `exchange.go` | Exchanges | `GetExchange`, `GetExchangeInstrument` |
| `instruments.go` | Instruments | `GetInstrumentMetaData`, `BatchGetInstrumentMetaData`, `ApiError` |
| `kline.go` | K-line / candlesticks | `GetKline`, `BatchGetKline`, `GetExFactor`, `Period`, `AdjustType`, `Kline`, `CompactKlineData` |
| `quote.go` | Quotes / market depth | `GetQuote`, `BatchGetQuote`, `GetDepth`, `BatchGetDepth`, `Quote`, `MarketDepth`, `Region`, `SessionStatus` |
| `universe.go` | Universes (symbol groups) | `GetUniverse`, `BatchGetUniverse`, `GetUniverseDetail` |
| `financial.go` | Financial statements | `GetBalanceSheet`, `GetCashFlow`, `GetIncome`, `GetMetrics`, `GetShare`, `FinancialReq` |

Each `*_test.go` file mirrors its source file.

### Client pattern

`TickFlow` holds `xApiKey` and `baseURL` (default `https://api.tickflow.org`). Every API method is a pointer-receiver on `*TickFlow`, takes `context.Context` as the first argument plus a request struct, and returns `(respPtr, error)`. Methods fall into two groups:

- **Singular** (e.g. `GetQuote`, `GetKline`, `GetExchangeInstrument`) — GET with query params, one symbol or a comma-joined `symbols` string.
- **Batch** (e.g. `BatchGetQuote`, `BatchGetInstrumentMetaData`, `BatchGetUniverse`) — POST with JSON body. Some accept `[]string` (POST body), others accept a comma-separated `string` (query param).

### Request validation convention

Methods validate inputs against package-level sentinel errors (`ErrNilReq`, `EmptyKey`, `ErrInvalidExchange`, `ErrEmptySymbol`, `ErrEmptySymbols`, `ErrTooManySymbols`, `ErrEmptyID`) and log via `log/slog` before returning. `MaxBatchSymbols = 1000` caps batch sizes. When adding or modifying a method, follow this validate → build URL → send → decode → return shape.

### Symbol format

Symbols are encoded as `"<code>.<exchange>"` (e.g. `600000.SH`, `AAPL.US`, `00700.HK`). Valid exchanges: `US`, `SH`, `SZ`, `BJ`, `HK` (see `validExchanges` in `tickflow.go`).

### K-line data shape

K-line responses use **columnar** `CompactKlineData` (parallel slices per field) rather than an array of per-bar objects — this is intentional for transfer efficiency. `ExFactor` entries carry a `Timestamp` + `ExFactor` for dividend/split adjustments.

## Testing Conventions

- Tests construct a `*TickFlow` directly with `baseURL` pointed at a local `httptest.NewServer` (`&TickFlow{xApiKey: "test-key", baseURL: ts.URL}`) — no real API calls.
- Each domain's test file defines a `setup<Domain>MockServer` helper that asserts method, path, `x-api-key` header, and query/body params, then returns a canned JSON response. Reuse the existing helper for that domain when adding tests.
- Use `testify/assert` and `testify/require`; structure cases as `t.Run("...", func(t *testing.T){...})` subtests.
- Cover: nil-request error path, validation errors, successful decode (assert every field), empty results, and server error responses (non-2xx → error returned).

## Workflow: Branches, PRs, and CHANGELOG

**Never commit directly to `main`.** The workflow for every change is:

1. Create a new branch from `main` (e.g. `feat/add-get-dividend`, `fix/kline-param`).
2. Implement the change and commit.
3. Update `CHANGELOG.md` (see below) — include this commit too.
4. Push the branch and open a PR against `main`.

### CHANGELOG format

`CHANGELOG.md` is versioned under `## vX.Y.Z` headings (newest on top). Each version contains a **numbered list** of changes written in Chinese, mirroring this existing style:

```markdown
## v0.0.3

1. 新增实时行情相关功能：`GetQuote` / `BatchGetQuote` / `GetDepth` / `BatchGetDepth`
2. 函数参数新增 context.Context
3. 新增标的池相关接口：`GetUniverse` / `BatchGetUniverse` / `GetUniverseDetail`
4. 新增 `GetExchange` 方法
```

Rules when adding an entry:
- Bump to the next version and insert a new `## vX.Y.Z` block at the **top** of the file.
- Use a numbered list (`1.`, `2.`, …); one item per logical change.
- Describe the change in Chinese, matching the existing tone (e.g. `新增 … 方法`, `新增 … 相关功能`, `修复 …`).
- Wrap Go method/type names in backticks, and group related items with ` / ` separators when it reads more cleanly (as in the v0.0.3 example).

## API Reference

Doc comments on each method link to the official docs (`https://docs.tickflow.org/zh-hans/api-reference/...`). The full OpenAPI spec is at `https://api.tickflow.org/openapi.json`.
