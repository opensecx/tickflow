package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// Period represents the time interval of a K-line (candlestick) bar.
type Period string

const (
	Period1m  Period = "1m"
	Period5m  Period = "5m"
	Period10m Period = "10m"
	Period15m Period = "15m"
	Period30m Period = "30m"
	Period60m Period = "60m"
	Period1d  Period = "1d"
	Period1w  Period = "1w"
	Period1M  Period = "1M"
	Period1Q  Period = "1Q"
	Period1Y  Period = "1Y"
)

// AdjustType represents the price adjustment type used for K-line data.
type AdjustType string

const (
	AdjustTypeForward          AdjustType = "forward"
	AdjustTypeBackward         AdjustType = "backward"
	AdjustTypeForwardAdditive  AdjustType = "forward_additive"
	AdjustTypeBackwardAdditive AdjustType = "backward_additive"
	AdjustTypeNone             AdjustType = "none"
)

// Kline represents a single K-line (candlestick) data point with OHLCV and optional fields.
type Kline struct {
	Timestamp       int64   `json:"timestamp"`        // 时间戳 (毫秒)
	Open            float64 `json:"open"`             // 开盘价
	High            float64 `json:"high"`             // 最高价
	Low             float64 `json:"low"`              // 最低价
	Close           float64 `json:"close"`            // 收盘价
	Volume          int64   `json:"volume"`           // 成交量
	Amount          float64 `json:"amount"`           // 成交额
	PrevClose       float64 `json:"prev_close"`       // 前收盘价 (可选)
	OpenInterest    float64 `json:"open_interest"`    // 持仓量 (可选)
	SettlementPrice float64 `json:"settlement_price"` // 结算价 (可选)
}

// CompactKlineData is a columnar (parallel-slice) representation of K-line data
// for efficient transfer. Each field is a slice where index i corresponds to bar i.
type CompactKlineData struct {
	Timestamp       []int64   `json:"timestamp"`        // 时间戳 (毫秒)
	Open            []float64 `json:"open"`             // 开盘价
	High            []float64 `json:"high"`             // 最高价
	Low             []float64 `json:"low"`              // 最低价
	Close           []float64 `json:"close"`            // 收盘价
	Volume          []int64   `json:"volume"`           // 成交量
	Amount          []float64 `json:"amount"`           // 成交额
	PrevClose       []float64 `json:"prev_close"`       // 前收盘价 (可选)
	OpenInterest    []float64 `json:"open_interest"`    // 持仓量 (可选)
	SettlementPrice []float64 `json:"settlement_price"` // 结算价 (可选)
}

// GetKlineReq is the request parameters for GetKline.
type GetKlineReq struct {
	Symbol    string     `json:"symbol"`               // 标的代码, 例如 "600000.SH"
	Period    Period     `json:"period,omitempty"`     // K线周期
	Count     int        `json:"count,omitempty"`      // 返回的K线数量 (默认100, 最大10000)
	StartTime int64      `json:"start_time,omitempty"` // 开始时间(毫秒时间戳)
	EndTime   int64      `json:"end_time,omitempty"`   // 结束时间(毫秒时间戳)
	Adjust    AdjustType `json:"adjust,omitempty"`     // 复权类型
}

// GetKlineResp is the response from GetKline.
type GetKlineResp struct {
	Data *CompactKlineData `json:"data"` // 紧凑列式K线数据
}

// GetKline returns K-line (candlestick) data for a single symbol.
//
// symbol is required, e.g. "600000.SH", "AAPL.US".
// period is optional; valid values: 1m, 5m, 10m, 15m, 30m, 60m, 1d, 1w, 1M, 1Q, 1Y.
// count is optional; default 100, maximum 10000.
// start_time / end_time are optional millisecond timestamps.
// adjust is optional: forward, backward, forward_additive, backward_additive, none.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/k%E7%BA%BF%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2-k%E7%BA%BF%E6%95%B0%E6%8D%AE
func (tf *TickFlow) GetKline(ctx context.Context, req *GetKlineReq) (resp *GetKlineResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbol == "" {
		err = ErrEmptySymbol
		slog.Error("[GetKline] empty symbol")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/klines", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbol", string(req.Symbol))

	if req.Period != "" {
		rb = rb.Param("period", string(req.Period))
	}
	if req.Count > 0 {
		rb = rb.ParamInt("count", req.Count)
	}
	if req.StartTime != 0 {
		rb = rb.ParamInt("start_time", int(req.StartTime))
	}
	if req.EndTime != 0 {
		rb = rb.ParamInt("end_time", int(req.EndTime))
	}
	if req.Adjust != "" {
		rb = rb.Param("adjust", string(req.Adjust))
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetKline] fail to request", "reqURL", reqURL, "symbol", req.Symbol, "error", err)
		return nil, err
	}

	return
}

// BatchGetKlineReq is the request parameters for BatchGetKline.
type BatchGetKlineReq struct {
	Symbols   string     `json:"symbols"`              // 标的代码，逗号分隔, 例如 "600000.SH,000001.SZ"
	Period    Period     `json:"period,omitempty"`     // K线周期
	Count     int        `json:"count,omitempty"`      // 返回的K线数量 (默认100, 最大10000)
	StartTime int64      `json:"start_time,omitempty"` // 开始时间(毫秒时间戳)
	EndTime   int64      `json:"end_time,omitempty"`   // 结束时间(毫秒时间戳)
	Adjust    AdjustType `json:"adjust,omitempty"`     // 复权类型
}

// BatchGetKlineResp is the response from BatchGetKline.
type BatchGetKlineResp struct {
	Data map[string]*CompactKlineData `json:"data"` // key 为标的代码，value 为紧凑列式K线数据
}

// BatchGetKline returns K-line data for multiple symbols in a single request.
//
// symbols is required, comma-separated, e.g. "600000.SH,000001.SZ".
// period is optional; valid values: 1m, 5m, 10m, 15m, 30m, 60m, 1d, 1w, 1M, 1Q, 1Y.
// count is optional; default 100, maximum 10000.
// start_time / end_time are optional millisecond timestamps.
// adjust is optional: forward, backward, forward_additive, backward_additive, none.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/k%E7%BA%BF%E6%95%B0%E6%8D%AE/%E6%89%B9%E9%87%8F%E6%9F%A5%E8%AF%A2-k%E7%BA%BF%E6%95%B0%E6%8D%AE
func (tf *TickFlow) BatchGetKline(ctx context.Context, req *BatchGetKlineReq) (resp *BatchGetKlineResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[BatchGetKline] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/klines/batch", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.Period != "" {
		rb = rb.Param("period", string(req.Period))
	}
	if req.Count > 0 {
		rb = rb.ParamInt("count", req.Count)
	}
	if req.StartTime != 0 {
		rb = rb.ParamInt("start_time", int(req.StartTime))
	}
	if req.EndTime != 0 {
		rb = rb.ParamInt("end_time", int(req.EndTime))
	}
	if req.Adjust != "" {
		rb = rb.Param("adjust", string(req.Adjust))
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[BatchGetKline] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}

// ExFactorEntry represents a single ex-factor (dividend/split adjustment) record.
type ExFactorEntry struct {
	Timestamp int64   `json:"timestamp"` // 除权日时间戳 (毫秒)
	ExFactor  float64 `json:"ex_factor"` // 除权因子
}

// GetExFactorReq is the request parameters for GetExFactor.
type GetExFactorReq struct {
	Symbols   string `json:"symbols"`              // 逗号分隔的标的代码, 例如 "600519.SH,000001.SZ"
	StartTime int64  `json:"start_time,omitempty"` // 开始时间(毫秒时间戳)
	EndTime   int64  `json:"end_time,omitempty"`   // 结束时间(毫秒时间戳)
}

// GetExFactorResp is the response from GetExFactor.
type GetExFactorResp struct {
	Data map[string][]ExFactorEntry `json:"data"` // key 为标的代码，value 为除权因子列表
}

// GetExFactor returns ex-factor (dividend/split adjustment) data for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_time / end_time are optional millisecond timestamps.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/k%E7%BA%BF%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E9%99%A4%E6%9D%83%E5%9B%A0%E5%AD%90
func (tf *TickFlow) GetExFactor(ctx context.Context, req *GetExFactorReq) (resp *GetExFactorResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetExFactor] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/klines/ex-factors", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartTime != 0 {
		rb = rb.ParamInt("start_time", int(req.StartTime))
	}
	if req.EndTime != 0 {
		rb = rb.ParamInt("end_time", int(req.EndTime))
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetExFactor] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}
