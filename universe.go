package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// ========== 标的池 ==========

// UniverseSummary is a summary of a universe (symbol group) used for list views.
type UniverseSummary struct {
	ID          string `json:"id"`           // 唯一标识符
	Name        string `json:"name"`         // 显示名称
	Category    string `json:"category"`     // 分类
	Region      string `json:"region"`       // 地区代码
	SymbolCount int    `json:"symbol_count"` // 标的数量
	Description string `json:"description"`  // 描述
}

// UniverseDetail is the full detail of a universe, including its symbol list.
type UniverseDetail struct {
	UniverseSummary
	Symbols []string `json:"symbols"` // 标的列表
}

// GetUniverseResp is the response from GetUniverse.
type GetUniverseResp struct {
	Data []UniverseSummary `json:"data"` // 标的池摘要列表
}

// GetUniverse returns the list of available universes (symbol groups).
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E6%A0%87%E7%9A%84%E6%B1%A0/%E8%8E%B7%E5%8F%96%E6%A0%87%E7%9A%84%E6%B1%A0%E5%88%97%E8%A1%A8
func (tf *TickFlow) GetUniverse(ctx context.Context) (resp *GetUniverseResp, err error) {
	reqURL := fmt.Sprintf("%s/v1/universes", tf.baseURL)
	err = requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[GetUniverse] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// BatchGetUniverseReq is the request parameters for BatchGetUniverse.
type BatchGetUniverseReq struct {
	IDs []string `json:"ids"` // 标的池 ID 列表，例如 ["CN_Equity_A", "CN_ETF"]
}

// BatchGetUniverseResp is the response from BatchGetUniverse.
type BatchGetUniverseResp struct {
	Data map[string]*UniverseDetail `json:"data"` // key 为标的池 ID
}

// BatchGetUniverse returns detailed information for a batch of universes.
//
// ids is the list of universe IDs, e.g. ["CN_Equity_A", "CN_ETF"].
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E6%A0%87%E7%9A%84%E6%B1%A0/%E6%89%B9%E9%87%8F%E8%8E%B7%E5%8F%96%E6%A0%87%E7%9A%84%E6%B1%A0%E8%AF%A6%E6%83%85
func (tf *TickFlow) BatchGetUniverse(ctx context.Context, req *BatchGetUniverseReq) (resp *BatchGetUniverseResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}

	reqURL := fmt.Sprintf("%s/v1/universes/batch", tf.baseURL)
	err = requests.
		URL(reqURL).
		Post().
		Header("x-api-key", tf.xApiKey).
		ContentType("application/json").
		BodyJSON(req).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[BatchGetUniverse] fail to request", "reqURL", reqURL, "error", err)
		return nil, err
	}

	return
}

// GetUniverseDetailResp is the response from GetUniverseDetail.
type GetUniverseDetailResp struct {
	Data *UniverseDetail `json:"data"` // 标的池详情
}

// GetUniverseDetail returns detailed information for a single universe.
//
// id is the universe ID, e.g. "CN_Equity_A".
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E6%A0%87%E7%9A%84%E6%B1%A0/%E8%8E%B7%E5%8F%96%E6%A0%87%E7%9A%84%E6%B1%A0%E8%AF%A6%E6%83%85
func (tf *TickFlow) GetUniverseDetail(ctx context.Context, id string) (resp *GetUniverseDetailResp, err error) {
	if id == "" {
		err = ErrEmptyID
		slog.Error("[GetUniverseDetail] empty id")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/universes/%s", tf.baseURL, id)
	err = requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		slog.Error("[GetUniverseDetail] fail to request", "reqURL", reqURL, "id", id, "error", err)
		return nil, err
	}

	return
}
