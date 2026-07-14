package tickflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBalanceSheet(t *testing.T) {
	req := &FinancialReq{Symbols: "600519.SH"}

	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetBalanceSheet(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetBalanceSheet(context.Background(), &FinancialReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetBalanceSheetResp{
			Data: map[string][]BalanceSheetRecord{
				"600519.SH": {
					{
						PeriodEnd:                  "2024-12-31",
						AnnounceDate:               "2025-03-31",
						TotalAssets:                25000000000,
						TotalCurrentAssets:         18000000000,
						TotalNonCurrentAssets:      7000000000,
						CashAndEquivalents:         5000000000,
						AccountsReceivable:         800000000,
						Inventory:                  3000000000,
						FixedAssets:                4500000000,
						IntangibleAssets:           500000000,
						Goodwill:                   200000000,
						TotalLiabilities:           10000000000,
						TotalCurrentLiabilities:    8000000000,
						TotalNonCurrentLiabilities: 2000000000,
						AccountsPayable:            1500000000,
						ShortTermBorrowing:         1000000000,
						LongTermBorrowing:          1500000000,
						TotalEquity:                15000000000,
						EquityAttributable:         14000000000,
						MinorityInterest:           1000000000,
						RetainedEarnings:           8000000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/balance-sheet", map[string]string{"symbols": "600519.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetBalanceSheet(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 1)

		records := resp.Data["600519.SH"]
		require.Len(t, records, 1)
		r := records[0]
		assert.Equal(t, "2024-12-31", r.PeriodEnd)
		assert.Equal(t, 25000000000.0, r.TotalAssets)
		assert.Equal(t, 5000000000.0, r.CashAndEquivalents)
		assert.Equal(t, 3000000000.0, r.Inventory)
		assert.Equal(t, 15000000000.0, r.TotalEquity)
	})

	t.Run("successful query with date range and latest", func(t *testing.T) {
		expectedResp := &GetBalanceSheetResp{
			Data: map[string][]BalanceSheetRecord{
				"AAPL.US": {
					{
						PeriodEnd:          "2024-12-31",
						TotalAssets:        350000000000,
						TotalCurrentAssets: 150000000000,
						TotalLiabilities:   280000000000,
						TotalEquity:        70000000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/balance-sheet", map[string]string{
			"symbols":    "AAPL.US",
			"start_date": "2024-01-01",
			"end_date":   "2024-12-31",
			"latest":     "true",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetBalanceSheet(context.Background(), &FinancialReq{
			Symbols:   "AAPL.US",
			StartDate: "2024-01-01",
			EndDate:   "2024-12-31",
			Latest:    true,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		records := resp.Data["AAPL.US"]
		require.Len(t, records, 1)
		assert.Equal(t, "2024-12-31", records[0].PeriodEnd)
	})

	t.Run("multiple symbols", func(t *testing.T) {
		symbols := "600519.SH,000001.SZ"
		expectedResp := &GetBalanceSheetResp{
			Data: map[string][]BalanceSheetRecord{
				"600519.SH": {
					{PeriodEnd: "2024-12-31", TotalAssets: 25000000000},
				},
				"000001.SZ": {
					{PeriodEnd: "2024-12-31", TotalAssets: 15000000000},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/balance-sheet", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetBalanceSheet(context.Background(), &FinancialReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 2)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/financials/balance-sheet", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetBalanceSheet(context.Background(), &FinancialReq{Symbols: "BAD"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "AUTH_FAILED",
				Message: "Invalid API key",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "bad-key", baseURL: ts.URL}
		resp, err := tf.GetBalanceSheet(context.Background(), &FinancialReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetCashFlow(t *testing.T) {
	req := &FinancialReq{Symbols: "600519.SH"}

	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetCashFlow(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetCashFlow(context.Background(), &FinancialReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetCashFlowResp{
			Data: map[string][]CashFlowRecord{
				"600519.SH": {
					{
						PeriodEnd:            "2024-12-31",
						AnnounceDate:         "2025-03-31",
						NetOperatingCashFlow: 8000000000,
						NetInvestingCashFlow: -3000000000,
						NetFinancingCashFlow: -2000000000,
						Capex:                -2500000000,
						NetCashChange:        3000000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/cash-flow", map[string]string{"symbols": "600519.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetCashFlow(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)

		records := resp.Data["600519.SH"]
		require.Len(t, records, 1)
		r := records[0]
		assert.Equal(t, "2024-12-31", r.PeriodEnd)
		assert.Equal(t, 8000000000.0, r.NetOperatingCashFlow)
		assert.Equal(t, -3000000000.0, r.NetInvestingCashFlow)
		assert.Equal(t, -2000000000.0, r.NetFinancingCashFlow)
		assert.Equal(t, -2500000000.0, r.Capex)
		assert.Equal(t, 3000000000.0, r.NetCashChange)
	})

	t.Run("successful query with latest", func(t *testing.T) {
		expectedResp := &GetCashFlowResp{
			Data: map[string][]CashFlowRecord{
				"AAPL.US": {
					{
						PeriodEnd:            "2024-12-31",
						NetOperatingCashFlow: 110000000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/cash-flow", map[string]string{
			"symbols": "AAPL.US",
			"latest":  "true",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetCashFlow(context.Background(), &FinancialReq{
			Symbols: "AAPL.US",
			Latest:  true,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		records := resp.Data["AAPL.US"]
		require.Len(t, records, 1)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/financials/cash-flow", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetCashFlow(context.Background(), &FinancialReq{Symbols: "BAD"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "AUTH_FAILED",
				Message: "Invalid API key",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "bad-key", baseURL: ts.URL}
		resp, err := tf.GetCashFlow(context.Background(), &FinancialReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetIncome(t *testing.T) {
	req := &FinancialReq{Symbols: "600519.SH"}

	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetIncome(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetIncome(context.Background(), &FinancialReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetIncomeResp{
			Data: map[string][]IncomeRecord{
				"600519.SH": {
					{
						PeriodEnd:             "2024-12-31",
						AnnounceDate:          "2025-03-31",
						Revenue:               150000000000,
						OperatingCost:         45000000000,
						OperatingProfit:       105000000000,
						TotalProfit:           80000000000,
						NetIncome:             75000000000,
						NetIncomeAttributable: 73000000000,
						NetIncomeDeducted:     72000000000,
						BasicEPS:              60.1,
						DilutedEPS:            60.0,
						SellingExpense:        5000000000,
						AdminExpense:          3000000000,
						FinancialExpense:      -500000000,
						RDExpense:             1000000000,
						IncomeTax:             5000000000,
						NonOperatingIncome:    100000000,
						NonOperatingExpense:   50000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/income", map[string]string{"symbols": "600519.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetIncome(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)

		records := resp.Data["600519.SH"]
		require.Len(t, records, 1)
		r := records[0]
		assert.Equal(t, "2024-12-31", r.PeriodEnd)
		assert.Equal(t, 150000000000.0, r.Revenue)
		assert.Equal(t, 75000000000.0, r.NetIncome)
		assert.Equal(t, 60.1, r.BasicEPS)
		assert.Equal(t, 73000000000.0, r.NetIncomeAttributable)
	})

	t.Run("successful query with date range", func(t *testing.T) {
		expectedResp := &GetIncomeResp{
			Data: map[string][]IncomeRecord{
				"AAPL.US": {
					{PeriodEnd: "2024-12-31", Revenue: 391000000000, NetIncome: 94000000000},
					{PeriodEnd: "2024-09-30", Revenue: 383000000000, NetIncome: 97000000000},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/income", map[string]string{
			"symbols":    "AAPL.US",
			"start_date": "2024-01-01",
			"end_date":   "2024-12-31",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetIncome(context.Background(), &FinancialReq{
			Symbols:   "AAPL.US",
			StartDate: "2024-01-01",
			EndDate:   "2024-12-31",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		records := resp.Data["AAPL.US"]
		assert.Len(t, records, 2)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/financials/income", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetIncome(context.Background(), &FinancialReq{Symbols: "BAD"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "AUTH_FAILED",
				Message: "Invalid API key",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "bad-key", baseURL: ts.URL}
		resp, err := tf.GetIncome(context.Background(), &FinancialReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetMetrics(t *testing.T) {
	req := &FinancialReq{Symbols: "600519.SH"}

	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetMetrics(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetMetrics(context.Background(), &FinancialReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetMetricsResp{
			Data: map[string][]MetricsRecord{
				"600519.SH": {
					{
						PeriodEnd:              "2024-12-31",
						AnnounceDate:           "2025-03-31",
						BPS:                    1200.5,
						EPSBasic:               60.1,
						EPSDiluted:             60.0,
						OCFPS:                  65.2,
						ROE:                    0.25,
						ROEDiluted:             0.248,
						ROA:                    0.15,
						GrossMargin:            0.92,
						NetMargin:              0.50,
						RevenueYoY:             0.18,
						NetIncomeYoY:           0.15,
						DebtToAssetRatio:       0.40,
						InventoryTurnover:      3.5,
						OperatingCashToRevenue: 0.55,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/metrics", map[string]string{"symbols": "600519.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetMetrics(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)

		records := resp.Data["600519.SH"]
		require.Len(t, records, 1)
		r := records[0]
		assert.Equal(t, "2024-12-31", r.PeriodEnd)
		assert.Equal(t, 1200.5, r.BPS)
		assert.Equal(t, 60.1, r.EPSBasic)
		assert.Equal(t, 0.25, r.ROE)
		assert.Equal(t, 0.50, r.NetMargin)
		assert.Equal(t, 0.18, r.RevenueYoY)
		assert.Equal(t, 0.40, r.DebtToAssetRatio)
	})

	t.Run("successful query with latest", func(t *testing.T) {
		expectedResp := &GetMetricsResp{
			Data: map[string][]MetricsRecord{
				"AAPL.US": {
					{
						PeriodEnd: "2024-12-31",
						ROE:       0.175,
						ROA:       0.125,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/metrics", map[string]string{
			"symbols": "AAPL.US",
			"latest":  "true",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetMetrics(context.Background(), &FinancialReq{
			Symbols: "AAPL.US",
			Latest:  true,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		records := resp.Data["AAPL.US"]
		require.Len(t, records, 1)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/financials/metrics", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetMetrics(context.Background(), &FinancialReq{Symbols: "BAD"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "AUTH_FAILED",
				Message: "Invalid API key",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "bad-key", baseURL: ts.URL}
		resp, err := tf.GetMetrics(context.Background(), &FinancialReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetShare(t *testing.T) {
	req := &FinancialReq{Symbols: "600519.SH"}

	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetShare(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetShare(context.Background(), &FinancialReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetShareResp{
			Data: map[string][]SharesRecord{
				"600519.SH": {
					{
						PeriodEnd:    "2024-12-31",
						AnnounceDate: "2025-03-31",
						TotalShares:  1256000000,
						FloatShares:  1256000000,
					},
					{
						PeriodEnd:    "2024-06-30",
						AnnounceDate: "2024-08-30",
						TotalShares:  1256000000,
						FloatShares:  1256000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/shares", map[string]string{"symbols": "600519.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetShare(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 1)

		records := resp.Data["600519.SH"]
		require.Len(t, records, 2)
		r := records[0]
		assert.Equal(t, "2024-12-31", r.PeriodEnd)
		assert.Equal(t, "2025-03-31", r.AnnounceDate)
		assert.Equal(t, 1256000000.0, r.TotalShares)
		assert.Equal(t, 1256000000.0, r.FloatShares)
	})

	t.Run("successful query with latest", func(t *testing.T) {
		expectedResp := &GetShareResp{
			Data: map[string][]SharesRecord{
				"AAPL.US": {
					{
						PeriodEnd:    "2024-12-31",
						AnnounceDate: "2025-01-31",
						TotalShares:  15400000000,
						FloatShares:  15300000000,
					},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/shares", map[string]string{
			"symbols": "AAPL.US",
			"latest":  "true",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetShare(context.Background(), &FinancialReq{
			Symbols: "AAPL.US",
			Latest:  true,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		records := resp.Data["AAPL.US"]
		require.Len(t, records, 1)
		assert.Equal(t, 15400000000.0, records[0].TotalShares)
		assert.Equal(t, 15300000000.0, records[0].FloatShares)
	})

	t.Run("multiple symbols", func(t *testing.T) {
		symbols := "600519.SH,00700.HK"
		expectedResp := &GetShareResp{
			Data: map[string][]SharesRecord{
				"600519.SH": {
					{PeriodEnd: "2024-12-31", TotalShares: 1256000000, FloatShares: 1256000000},
				},
				"00700.HK": {
					{PeriodEnd: "2024-12-31", TotalShares: 9420000000, FloatShares: 9200000000},
				},
			},
		}

		ts := setupFinancialMockServer(t, "/v1/financials/shares", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetShare(context.Background(), &FinancialReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 2)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/financials/shares", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetShare(context.Background(), &FinancialReq{Symbols: "BAD"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "AUTH_FAILED",
				Message: "Invalid API key",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "bad-key", baseURL: ts.URL}
		resp, err := tf.GetShare(context.Background(), &FinancialReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// setupFinancialMockServer 创建模拟 HTTP GET 财务数据服务器，验证请求路径和查询参数
func setupFinancialMockServer(t *testing.T, expectedPath string, expectedParams map[string]string, respBody any) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, http.MethodGet, r.Method)

		// 验证请求路径
		assert.Equal(t, expectedPath, r.URL.Path)

		// 验证 x-api-key header
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

		// 验证查询参数（不考虑顺序，自动处理 URL 编码）
		query := r.URL.Query()
		for key, expectedVal := range expectedParams {
			actualVal := query.Get(key)
			assert.Equal(t, expectedVal, actualVal, "query param %s mismatch", key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(respBody)
		assert.Nil(t, err)
	}))
}
