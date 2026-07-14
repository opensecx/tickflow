package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// ApiError API 错误响应
type ApiError struct {
	Code    string `json:"code"`    // 错误码
	Message string `json:"message"` // 错误描述
	Details string `json:"details"` // 可选调试信息
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("tickflow api error [%s]: %s", e.Code, e.Message)
}

// isValidExchange 校验交易所代码是否合法
func isValidExchange(exchange string) bool {
	return validExchanges[exchange]
}

// GetInstrumentMetaDataReq 获取标的元数据请求
type GetInstrumentMetaDataReq struct {
	Symbols string `json:"symbols"` // 标的代码 "TSLA.US,600036.SH"
}

// GetInstrumentMetaDataResp 获取标的元数据响应
type GetInstrumentMetaDataResp struct {
	Exchange string       `json:"exchange"` // 交易所代码
	Count    int          `json:"count"`    // 标的数量
	Data     []Instrument `json:"data"`     // 标的列表
}

// GetInstrumentMetaData 获取标的元数据
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E6%A0%87%E7%9A%84/%E6%9F%A5%E8%AF%A2%E6%A0%87%E7%9A%84%E5%85%83%E6%95%B0%E6%8D%AE
// exchange 可选值: US, SH, SZ, BJ, HK
// instrumentType 可选，用于按类型过滤，传空字符串表示不过滤
func (tf *TickFlow) GetInstrumentMetaData(ctx context.Context, req *GetInstrumentMetaDataReq) (resp *GetInstrumentMetaDataResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}

	reqURL := fmt.Sprintf("%s/v1/instruments?symbols=%s", tf.baseURL, req.Symbols)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey)

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetInstrumentMetaData] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// BatchGetInstrumentMetaDataReq 批量查询标的元数据请求
type BatchGetInstrumentMetaDataReq struct {
	Symbols []string `json:"symbols"` // 标的代码列表，最多 1000 个
}

// BatchGetInstrumentMetaDataResp 批量查询标的元数据响应
type BatchGetInstrumentMetaDataResp struct {
	Data []Instrument `json:"data"` // 标的元数据列表
}

// BatchGetInstrumentMetaData 批量查询标的元数据
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E6%A0%87%E7%9A%84/%E6%89%B9%E9%87%8F%E6%9F%A5%E8%AF%A2%E6%A0%87%E7%9A%84%E5%85%83%E6%95%B0%E6%8D%AE
// symbols 标的代码列表，格式为 "代码.交易所"（如 "600000.SH"、"AAPL.US"），最多 1000 个
func (tf *TickFlow) BatchGetInstrumentMetaData(ctx context.Context, req *BatchGetInstrumentMetaDataReq) (resp *BatchGetInstrumentMetaDataResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}

	if len(req.Symbols) == 0 {
		err = ErrEmptySymbols
		slog.Error("[BatchGetInstrumentMetaData] empty symbols")
		return
	}

	if len(req.Symbols) > MaxBatchSymbols {
		err = ErrTooManySymbols
		slog.Error("[BatchGetInstrumentMetaData] too many symbols", "num", len(req.Symbols))
		return
	}

	reqURL := fmt.Sprintf("%s/v1/instruments", tf.baseURL)
	err = requests.
		URL(reqURL).
		Post().
		Header("x-api-key", tf.xApiKey).
		ContentType("application/json").
		BodyJSON(req).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[BatchGetInstrumentMetaData] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}
