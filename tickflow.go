package tickflow

import (
	"errors"
)

const (
	// 默认 api 地址
	defaultBaseURL = "https://api.tickflow.org"
	// MaxBatchSymbols 批量查询最大标的数量
	MaxBatchSymbols = 1000
)

var (
	ErrNilReq          = errors.New("nil req")
	ErrEmptyKey        = errors.New("empty key")
	ErrInvalidExchange = errors.New("invalid exchange, must be one of: US, SH, SZ, BJ, HK")
	ErrEmptySymbols    = errors.New("symbols must not be empty")
	ErrTooManySymbols  = errors.New("symbols must not exceed 1000")
)

// validExchanges 支持的交易所列表
var validExchanges = map[string]bool{
	"US": true,
	"SH": true,
	"SZ": true,
	"BJ": true,
	"HK": true,
}

// TickFlow tickflow API 客户端
type TickFlow struct {
	xApiKey string // http 鉴权key
	baseURL string // API 基础地址
}

// NewTickFlow 创建 TickFlow 客户端
func NewTickFlow(key string) (*TickFlow, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}

	return &TickFlow{
		xApiKey: key,
		baseURL: defaultBaseURL,
	}, nil
}
