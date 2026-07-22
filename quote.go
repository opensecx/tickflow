package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// Region represents a supported geographic market.
type Region string

const (
	RegionCN Region = "CN"
	RegionUS Region = "US"
	RegionHK Region = "HK"
)

// SessionStatus represents the current trading session status of a market.
type SessionStatus string

const (
	SessionPreMarket  SessionStatus = "pre_market"
	SessionRegular    SessionStatus = "regular"
	SessionAfterHours SessionStatus = "after_hours"
	SessionClosed     SessionStatus = "closed"
	SessionHalted     SessionStatus = "halted"
	SessionLunchBreak SessionStatus = "lunch_break"
)

// QuoteExtension contains region-specific extension data for a quote.
type QuoteExtension struct {
	Type string `json:"type"` // 市场类型标签
}

// Quote represents a real-time market quote for an instrument.
type Quote struct {
	Symbol    string          `json:"symbol"`     // 标的代码
	Timestamp int64           `json:"timestamp"`  // 时间戳 (毫秒)
	LastPrice float64         `json:"last_price"` // 最新价
	Open      float64         `json:"open"`       // 开盘价
	High      float64         `json:"high"`       // 最高价
	Low       float64         `json:"low"`        // 最低价
	PrevClose float64         `json:"prev_close"` // 昨收价
	Volume    int64           `json:"volume"`     // 成交量
	Amount    float64         `json:"amount"`     // 成交额
	Region    Region          `json:"region"`     // 地区
	Session   SessionStatus   `json:"session"`    // 交易时段
	Ext       *QuoteExtension `json:"ext"`        // 扩展数据
}

// GetQuoteReq is the request parameters for GetQuote.
type GetQuoteReq struct {
	Symbols   string `json:"symbols"`   // 标的代码，逗号分隔，例如 "600000.SH,000001.SZ"
	Universes string `json:"universes"` // 标的池 ID，逗号分隔，例如 "CN_Equity_A,CN_ETF"
}

// GetQuoteResp is the response from GetQuote.
type GetQuoteResp struct {
	Data []Quote `json:"data"` // 行情数据列表
}

// GetQuote returns real-time quotes. Either symbols or universes may be provided.
//
// symbols is optional, comma-separated, e.g. "600000.SH,000001.SZ".
// universes is optional, comma-separated, e.g. "CN_Equity_A,CN_ETF".
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85/%E6%9F%A5%E8%AF%A2%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85
func (tf *TickFlow) GetQuote(ctx context.Context, req *GetQuoteReq) (resp *GetQuoteResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}

	reqURL := fmt.Sprintf("%s/v1/quotes", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey)

	if req.Symbols != "" {
		rb = rb.Param("symbols", req.Symbols)
	}
	if req.Universes != "" {
		rb = rb.Param("universes", req.Universes)
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetQuote] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// BatchGetQuoteReq is the request parameters for BatchGetQuote.
type BatchGetQuoteReq struct {
	Symbols   []string `json:"symbols"`   // 标的代码列表，例如 ["600000.SH", "000001.SZ", "AAPL.US"]
	Universes []string `json:"universes"` // 标的池 ID 列表，例如 ["CN_Equity_A", "CN_ETF"]
}

// BatchGetQuoteResp is the response from BatchGetQuote.
type BatchGetQuoteResp struct {
	Data []Quote `json:"data"` // 行情数据列表
}

// BatchGetQuote returns real-time quotes for a batch of symbols or universes.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85/%E6%89%B9%E9%87%8F%E6%9F%A5%E8%AF%A2%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85
func (tf *TickFlow) BatchGetQuote(ctx context.Context, req *BatchGetQuoteReq) (resp *BatchGetQuoteResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}

	reqURL := fmt.Sprintf("%s/v1/quotes", tf.baseURL)
	err = requests.
		URL(reqURL).
		Post().
		Header("x-api-key", tf.xApiKey).
		ContentType("application/json").
		BodyJSON(req).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[BatchGetQuote] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// ========== 市场深度（五档行情） ==========

// MarketDepth represents market depth (level 1, five-level bid/ask) data.
type MarketDepth struct {
	Symbol     string    `json:"symbol"`      // 标的代码
	Timestamp  int64     `json:"timestamp"`   // 时间戳 (毫秒)
	BidPrices  []float64 `json:"bid_prices"`  // 买入价 (买1-买5，降序)
	BidVolumes []int64   `json:"bid_volumes"` // 买入量
	AskPrices  []float64 `json:"ask_prices"`  // 卖出价 (卖1-卖5，升序)
	AskVolumes []int64   `json:"ask_volumes"` // 卖出量
	Region     Region    `json:"region"`      // 地区
}

// GetDepthReq is the request parameters for GetDepth.
type GetDepthReq struct {
	Symbol string `json:"symbol"` // 标的代码，例如 "600000.SH"
}

// GetDepthResp is the response from GetDepth.
type GetDepthResp struct {
	Data *MarketDepth `json:"data"` // 市场深度
}

// GetDepth returns market depth (five-level bid/ask) for a single symbol.
//
// symbol is required, e.g. "600000.SH".
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85/%E6%9F%A5%E8%AF%A2%E5%B8%82%E5%9C%BA%E6%B7%B1%E5%BA%A6%EF%BC%88%E4%BA%94%E6%A1%A3%E8%A1%8C%E6%83%85%EF%BC%89
func (tf *TickFlow) GetDepth(ctx context.Context, req *GetDepthReq) (resp *GetDepthResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbol == "" {
		err = ErrEmptySymbol
		slog.Error("[GetDepth] empty symbol")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/depth", tf.baseURL)
	err = requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbol", req.Symbol).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[GetDepth] fail to request", "reqURL", reqURL, "symbol", req.Symbol, "error", err)
		return nil, err
	}

	return
}

// BatchGetDepthReq is the request parameters for BatchGetDepth.
type BatchGetDepthReq struct {
	Symbols string `json:"symbols"` // 逗号分隔的标的代码，例如 "600000.SH,000001.SZ"
}

// BatchGetDepthResp is the response from BatchGetDepth.
type BatchGetDepthResp struct {
	Data map[string]*MarketDepth `json:"data"` // key 为标的代码
}

// BatchGetDepth returns market depth (five-level bid/ask) for multiple symbols.
//
// symbols is required, comma-separated, e.g. "600000.SH,000001.SZ".
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E5%AE%9E%E6%97%B6%E8%A1%8C%E6%83%85/%E6%89%B9%E9%87%8F%E6%9F%A5%E8%AF%A2%E5%B8%82%E5%9C%BA%E6%B7%B1%E5%BA%A6%EF%BC%88%E4%BA%94%E6%A1%A3%E8%A1%8C%E6%83%85%EF%BC%89
func (tf *TickFlow) BatchGetDepth(ctx context.Context, req *BatchGetDepthReq) (resp *BatchGetDepthResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[BatchGetDepth] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/depth/batch", tf.baseURL)
	err = requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[BatchGetDepth] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}
