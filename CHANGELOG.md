# TickFlow Changelog

## v0.0.9

1. 调整 K 线请求结构体中 `StartTime` 和 `EndTime` 字段类型：由 `*int64` 改为 `int64`（`GetKlineReq` / `BatchGetKlineReq` / `GetExFactorReq`）
2. 同步修复 `kline_test.go` 中的用例并移除不再需要的 `int64Ptr` 辅助函数

## v0.0.8

1. 调整 `Instrument` 中 `FloatShares` 和 `TotalShares` 为 float64 类型

## v0.0.7

1. 调整 `Instrument` 结构体字段类型：`Name` 由 `*string` 改为 `string`，`Type` 由 `*InstrumentType` 改为 `InstrumentType`
2. 调整 `GetExchangeInstrumentResp.Data` 字段类型：由 `[]Instrument` 改为 `[]*Instrument`
3. 同步修复 `exchange_test.go` 与 `instruments_test.go` 中的用例以匹配上述类型变更

## v0.0.6

1. 补充包级别注释及所有导出类型/方法/常量的 godoc 注释，遵循 Go 官方文档规范，确保 pkg.go.dev 正确渲染

## v0.0.5

1. 补充单元测试覆盖缺失分支，整体语句覆盖率提升至 100%（`ApiError.Error` / `GetCashFlow` / `GetIncome` / `GetMetrics` / `GetShare` / `BatchGetKline`）

## v0.0.4

1. 新增 `CLAUDE.md` 仓库工作流与架构说明文档

## v0.0.3

1. 新增实时行情相关功能：`GetQuote` / `BatchGetQuote` / `GetDepth` / `BatchGetDepth`
2. 函数参数新增 context.Context
3. 新增标的池相关接口：`GetUniverse` / `BatchGetUniverse` / `GetUniverseDetail`
4. 新增 `GetExchange` 方法

## v0.0.2

1. 新增 .github action
2. 实现 kline 相关方法：`GetKline` / `BatchGetKline` / `GetExFactor`
3. 新增 golangci action
4. 新增财务数据相关方法：`GetBalanceSheet` / `GetCashFlow` / `GetIncome` / `GetMetrics` / `GetShare`

## v0.0.1

1. 初始化 go 代码
2. 实现 `GetExchangeInstrument` `GetInstrumentMetaData` `BatchGetInstrumentMetaData` 方法
