package tickflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/carlmjohnson/requests"
)

// ========== 公共请求类型 ==========

// FinancialReq contains the common request parameters for financial statement queries.
type FinancialReq struct {
	Symbols   string `json:"symbols"`              // 逗号分隔的标的代码，如 "600519.SH,000001.SZ"
	StartDate string `json:"start_date,omitempty"` // 起始日期（可选，YYYY-MM-DD），筛选 period_end >= start_date
	EndDate   string `json:"end_date,omitempty"`   // 截止日期（可选，YYYY-MM-DD），筛选 period_end <= end_date
	Latest    bool   `json:"latest,omitempty"`     // 仅返回最新一期（默认 false）
}

// ========== 资产负债表 ==========

// BalanceSheetRecord represents a single-period balance sheet record.
type BalanceSheetRecord struct {
	PeriodEnd                  string  `json:"period_end"`                    // 报告期末日期 (YYYY-MM-DD)
	AnnounceDate               string  `json:"announce_date"`                 // 公告日期 (YYYY-MM-DD)
	TotalAssets                float64 `json:"total_assets"`                  // 资产总计
	TotalCurrentAssets         float64 `json:"total_current_assets"`          // 流动资产合计
	TotalNonCurrentAssets      float64 `json:"total_non_current_assets"`      // 非流动资产合计
	CashAndEquivalents         float64 `json:"cash_and_equivalents"`          // 货币资金
	AccountsReceivable         float64 `json:"accounts_receivable"`           // 应收账款
	Inventory                  float64 `json:"inventory"`                     // 存货
	FixedAssets                float64 `json:"fixed_assets"`                  // 固定资产
	IntangibleAssets           float64 `json:"intangible_assets"`             // 无形资产
	Goodwill                   float64 `json:"goodwill"`                      // 商誉
	TotalLiabilities           float64 `json:"total_liabilities"`             // 负债合计
	TotalCurrentLiabilities    float64 `json:"total_current_liabilities"`     // 流动负债合计
	TotalNonCurrentLiabilities float64 `json:"total_non_current_liabilities"` // 非流动负债合计
	AccountsPayable            float64 `json:"accounts_payable"`              // 应付账款
	ShortTermBorrowing         float64 `json:"short_term_borrowing"`          // 短期借款
	LongTermBorrowing          float64 `json:"long_term_borrowing"`           // 长期借款
	TotalEquity                float64 `json:"total_equity"`                  // 所有者权益合计
	EquityAttributable         float64 `json:"equity_attributable"`           // 归属于母公司股东权益
	MinorityInterest           float64 `json:"minority_interest"`             // 少数股东权益
	RetainedEarnings           float64 `json:"retained_earnings"`             // 未分配利润
}

// GetBalanceSheetResp is the response from GetBalanceSheet.
type GetBalanceSheetResp struct {
	Data map[string][]BalanceSheetRecord `json:"data"` // key 为标的代码
}

// GetBalanceSheet returns balance sheet data for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_date / end_date are optional, YYYY-MM-DD.
// latest is optional; when true, only the most recent period is returned.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E8%B5%84%E4%BA%A7%E8%B4%9F%E5%80%BA%E8%A1%A8
func (tf *TickFlow) GetBalanceSheet(ctx context.Context, req *FinancialReq) (resp *GetBalanceSheetResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetBalanceSheet] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/financials/balance-sheet", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartDate != "" {
		rb = rb.Param("start_date", req.StartDate)
	}
	if req.EndDate != "" {
		rb = rb.Param("end_date", req.EndDate)
	}
	if req.Latest {
		rb = rb.Param("latest", "true")
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetBalanceSheet] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}

// ========== 现金流量表 ==========

// CashFlowRecord represents a single-period cash flow record.
type CashFlowRecord struct {
	PeriodEnd            string  `json:"period_end"`              // 报告期末日期 (YYYY-MM-DD)
	AnnounceDate         string  `json:"announce_date"`           // 公告日期 (YYYY-MM-DD)
	NetOperatingCashFlow float64 `json:"net_operating_cash_flow"` // 经营活动现金流量净额
	NetInvestingCashFlow float64 `json:"net_investing_cash_flow"` // 投资活动现金流量净额
	NetFinancingCashFlow float64 `json:"net_financing_cash_flow"` // 筹资活动现金流量净额
	Capex                float64 `json:"capex"`                   // 资本支出
	NetCashChange        float64 `json:"net_cash_change"`         // 现金及现金等价物净增加额
}

// GetCashFlowResp is the response from GetCashFlow.
type GetCashFlowResp struct {
	Data map[string][]CashFlowRecord `json:"data"` // key 为标的代码
}

// GetCashFlow returns cash flow statement data for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_date / end_date are optional, YYYY-MM-DD.
// latest is optional; when true, only the most recent period is returned.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E7%8E%B0%E9%87%91%E6%B5%81%E9%87%8F%E8%A1%A8
func (tf *TickFlow) GetCashFlow(ctx context.Context, req *FinancialReq) (resp *GetCashFlowResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetCashFlow] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/financials/cash-flow", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartDate != "" {
		rb = rb.Param("start_date", req.StartDate)
	}
	if req.EndDate != "" {
		rb = rb.Param("end_date", req.EndDate)
	}
	if req.Latest {
		rb = rb.Param("latest", "true")
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetCashFlow] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}

// ========== 利润表 ==========

// IncomeRecord represents a single-period income statement record.
type IncomeRecord struct {
	PeriodEnd             string  `json:"period_end"`              // 报告期末日期 (YYYY-MM-DD)
	AnnounceDate          string  `json:"announce_date"`           // 公告日期 (YYYY-MM-DD)
	Revenue               float64 `json:"revenue"`                 // 营业总收入
	OperatingCost         float64 `json:"operating_cost"`          // 营业总成本
	OperatingProfit       float64 `json:"operating_profit"`        // 营业利润
	TotalProfit           float64 `json:"total_profit"`            // 利润总额
	NetIncome             float64 `json:"net_income"`              // 净利润
	NetIncomeAttributable float64 `json:"net_income_attributable"` // 归属于母公司股东的净利润
	NetIncomeDeducted     float64 `json:"net_income_deducted"`     // 扣除非经常性损益后的净利润
	BasicEPS              float64 `json:"basic_eps"`               // 基本每股收益
	DilutedEPS            float64 `json:"diluted_eps"`             // 稀释每股收益
	SellingExpense        float64 `json:"selling_expense"`         // 销售费用
	AdminExpense          float64 `json:"admin_expense"`           // 管理费用
	FinancialExpense      float64 `json:"financial_expense"`       // 财务费用
	RDExpense             float64 `json:"rd_expense"`              // 研发费用
	IncomeTax             float64 `json:"income_tax"`              // 所得税费用
	NonOperatingIncome    float64 `json:"non_operating_income"`    // 营业外收入
	NonOperatingExpense   float64 `json:"non_operating_expense"`   // 营业外支出
}

// GetIncomeResp is the response from GetIncome.
type GetIncomeResp struct {
	Data map[string][]IncomeRecord `json:"data"` // key 为标的代码
}

// GetIncome returns income statement data for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_date / end_date are optional, YYYY-MM-DD.
// latest is optional; when true, only the most recent period is returned.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E5%88%A9%E6%B6%A6%E8%A1%A8
func (tf *TickFlow) GetIncome(ctx context.Context, req *FinancialReq) (resp *GetIncomeResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetIncome] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/financials/income", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartDate != "" {
		rb = rb.Param("start_date", req.StartDate)
	}
	if req.EndDate != "" {
		rb = rb.Param("end_date", req.EndDate)
	}
	if req.Latest {
		rb = rb.Param("latest", "true")
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetIncome] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}

// ========== 核心财务指标 ==========

// MetricsRecord represents a single-period metrics record covering per-share data,
// profitability, growth, and solvency.
type MetricsRecord struct {
	PeriodEnd              string  `json:"period_end"`                // 报告期末日期 (YYYY-MM-DD)
	AnnounceDate           string  `json:"announce_date"`             // 公告日期 (YYYY-MM-DD)
	BPS                    float64 `json:"bps"`                       // 每股净资产
	EPSBasic               float64 `json:"eps_basic"`                 // 基本每股收益
	EPSDiluted             float64 `json:"eps_diluted"`               // 稀释每股收益
	OCFPS                  float64 `json:"ocfps"`                     // 每股经营现金流
	ROE                    float64 `json:"roe"`                       // 净资产收益率
	ROEDiluted             float64 `json:"roe_diluted"`               // 摊薄净资产收益率
	ROA                    float64 `json:"roa"`                       // 总资产收益率
	GrossMargin            float64 `json:"gross_margin"`              // 毛利率
	NetMargin              float64 `json:"net_margin"`                // 净利率
	RevenueYoY             float64 `json:"revenue_yoy"`               // 营收同比增长率
	NetIncomeYoY           float64 `json:"net_income_yoy"`            // 净利润同比增长率
	DebtToAssetRatio       float64 `json:"debt_to_asset_ratio"`       // 资产负债率
	InventoryTurnover      float64 `json:"inventory_turnover"`        // 存货周转率
	OperatingCashToRevenue float64 `json:"operating_cash_to_revenue"` // 销售现金比率
}

// GetMetricsResp is the response from GetMetrics.
type GetMetricsResp struct {
	Data map[string][]MetricsRecord `json:"data"` // key 为标的代码
}

// GetMetrics returns core financial metrics for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_date / end_date are optional, YYYY-MM-DD.
// latest is optional; when true, only the most recent period is returned.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E6%A0%B8%E5%BF%83%E8%B4%A2%E5%8A%A1%E6%8C%87%E6%A0%87
func (tf *TickFlow) GetMetrics(ctx context.Context, req *FinancialReq) (resp *GetMetricsResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetMetrics] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/financials/metrics", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartDate != "" {
		rb = rb.Param("start_date", req.StartDate)
	}
	if req.EndDate != "" {
		rb = rb.Param("end_date", req.EndDate)
	}
	if req.Latest {
		rb = rb.Param("latest", "true")
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetMetrics] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}

// ========== 股本表 ==========

// SharesRecord represents a single-period share structure record.
type SharesRecord struct {
	PeriodEnd    string  `json:"period_end"`    // 报告期末日期 (YYYY-MM-DD)
	AnnounceDate string  `json:"announce_date"` // 公告日期 (YYYY-MM-DD)
	TotalShares  float64 `json:"total_shares"`  // 总股本
	FloatShares  float64 `json:"float_shares"`  // 流通股本
}

// GetShareResp is the response from GetShare.
type GetShareResp struct {
	Data map[string][]SharesRecord `json:"data"` // key 为标的代码
}

// GetShare returns share structure data for one or more symbols.
//
// symbols is required, comma-separated, e.g. "600519.SH,000001.SZ".
// start_date / end_date are optional, YYYY-MM-DD.
// latest is optional; when true, only the most recent period is returned.
//
// api-url: https://docs.tickflow.org/zh-hans/api-reference/%E8%B4%A2%E5%8A%A1%E6%95%B0%E6%8D%AE/%E6%9F%A5%E8%AF%A2%E8%82%A1%E6%9C%AC%E8%A1%A8
func (tf *TickFlow) GetShare(ctx context.Context, req *FinancialReq) (resp *GetShareResp, err error) {
	if req == nil {
		return nil, ErrNilReq
	}
	if req.Symbols == "" {
		err = ErrEmptySymbols
		slog.Error("[GetShare] empty symbols")
		return
	}

	reqURL := fmt.Sprintf("%s/v1/financials/shares", tf.baseURL)
	rb := requests.
		URL(reqURL).
		Header("x-api-key", tf.xApiKey).
		Param("symbols", req.Symbols)

	if req.StartDate != "" {
		rb = rb.Param("start_date", req.StartDate)
	}
	if req.EndDate != "" {
		rb = rb.Param("end_date", req.EndDate)
	}
	if req.Latest {
		rb = rb.Param("latest", "true")
	}

	err = rb.ToJSON(&resp).Fetch(ctx)
	if err != nil {
		slog.Error("[GetShare] fail to request", "reqURL", reqURL, "symbols", req.Symbols, "error", err)
		return nil, err
	}

	return
}
