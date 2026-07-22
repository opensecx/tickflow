# TickFlow Changelog

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
