package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// GetExchangeInstrumentReq is the request parameters for GetExchangeInstrument.
type GetExchangeInstrumentReq struct {
	Exchange string `json:"Exchange"` // 交易所名称
	Type     string `json:"type"`     // 标的类型
}

// GetExchangeInstrumentResp is the response from GetExchangeInstrument.
type GetExchangeInstrumentResp struct {
	Exchange string        `json:"exchange"` // 交易所代码
	Count    int           `json:"count"`    // 标的数量
	Data     []*Instrument `json:"data"`     // 标的列表
}

// GetExchangeInstrument returns the list of instruments for a given exchange.
//
// exchange must be one of: US, SH, SZ, BJ, HK.
// instrumentType is optional; pass an empty string to return all types.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E4%BA%A4%E6%98%93%E6%89%80/%E8%8E%B7%E5%8F%96%E4%BA%A4%E6%98%93%E6%89%80%E7%9A%84%E6%A0%87%E7%9A%84%E5%88%97%E8%A1%A8
func (tf *TickFlow) GetExchangeInstrument(ctx context.Context, req *GetExchangeInstrumentReq) (resp *GetExchangeInstrumentResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if !isValidExchange(req.Exchange) {
		err = ErrInvalidExchange
		slog.Error("[GetExchangeInstrument] invalid param", "exchange", req.Exchange)
		return
	}

	reqURL := fmt.Sprintf("%s/v1/exchanges/%s/instruments", tf.baseURL, req.Exchange)

	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey)

	// 如果指定了标的类型，添加查询参数
	if req.Type != "" {
		rb = rb.Param("type", req.Type)
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetExchangeInstrument] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// ========== 交易所 ==========

// ExchangeInfo is a summary of an exchange.
type ExchangeInfo struct {
	Exchange string `json:"exchange"` // 交易所代码
	Region   string `json:"region"`   // 所属地区
	Count    int    `json:"count"`    // 标的数量
}

// GetExchangeResp is the response from GetExchange.
type GetExchangeResp struct {
	Data []ExchangeInfo `json:"data"` // 交易所列表
}

// GetExchange returns the list of supported exchanges.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E4%BA%A4%E6%98%93%E6%89%80/%E8%8E%B7%E5%8F%96%E4%BA%A4%E6%98%93%E6%89%80%E5%88%97%E8%A1%A8
func (tf *TickFlow) GetExchange(ctx context.Context) (resp *GetExchangeResp, err error) {
	reqURL := fmt.Sprintf("%s/v1/exchanges", tf.baseURL)
	err = requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[GetExchange] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}
