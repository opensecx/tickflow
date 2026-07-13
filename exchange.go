package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// GetExchangeInstrumentReq 获取交易所标的列表请求
type GetExchangeInstrumentReq struct {
	Exchange string `json:"Exchange"` // 交易所名称
	Type     string `json:"type"`     // 标的类型
}

// GetExchangeInstrumentResp 获取交易所标的列表响应
type GetExchangeInstrumentResp struct {
	Exchange string       `json:"exchange"` // 交易所代码
	Count    int          `json:"count"`    // 标的数量
	Data     []Instrument `json:"data"`     // 标的列表
}

// GetExchangeInstrument 获取交易所标的
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E4%BA%A4%E6%98%93%E6%89%80/%E8%8E%B7%E5%8F%96%E4%BA%A4%E6%98%93%E6%89%80%E7%9A%84%E6%A0%87%E7%9A%84%E5%88%97%E8%A1%A8
// exchange 可选值: US, SH, SZ, BJ, HK
// instrumentType 可选，用于按类型过滤，传空字符串表示不过滤
func (tf *TickFlow) GetExchangeInstrument(req *GetExchangeInstrumentReq) (resp *GetExchangeInstrumentResp, err error) {
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

	err = rb.ToJSON(&resp).Fetch(context.Background())
	if err != nil {
		slog.Error("[GetExchangeInstrument] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}
