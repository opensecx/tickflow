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

func TestIsValidExchange(t *testing.T) {
	tests := []struct {
		exchange string
		valid    bool
	}{
		{"US", true},
		{"SH", true},
		{"SZ", true},
		{"BJ", true},
		{"HK", true},
		{"us", false},
		{"sh", false},
		{"XX", false},
		{"", false},
		{"NASDAQ", false},
	}

	for _, tc := range tests {
		t.Run(tc.exchange, func(t *testing.T) {
			assert.Equal(t, tc.valid, isValidExchange(tc.exchange))
		})
	}
}

func TestApiError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		err := &ApiError{
			Code:    "INVALID_SYMBOL",
			Message: "Invalid symbol format: BAD",
		}
		assert.Equal(t, "tickflow api error [INVALID_SYMBOL]: Invalid symbol format: BAD", err.Error())
	})

	t.Run("empty code and message", func(t *testing.T) {
		err := &ApiError{}
		assert.Equal(t, "tickflow api error []: ", err.Error())
	})
}

func TestGetInstrumentMetaData(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols", func(t *testing.T) {
		expectedResp := &GetInstrumentMetaDataResp{
			Exchange: "US",
			Count:    0,
			Data:     []Instrument{},
		}

		ts := setupMockServer(t, "/v1/instruments", "symbols=", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: ""})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 0, resp.Count)
		assert.Empty(t, resp.Data)
	})

	t.Run("single symbol", func(t *testing.T) {
		expectedResp := &GetInstrumentMetaDataResp{
			Exchange: "US",
			Count:    1,
			Data: []Instrument{
				{
					Symbol:   "AAPL.US",
					Exchange: "US",
					Code:     "AAPL",
					Region:   "US",
					Name:     "Apple Inc.",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "us_equity",
						FloatShares: 15400000000,
						TotalShares: 15400000000,
					},
				},
			},
		}

		ts := setupMockServer(t, "/v1/instruments", "symbols=AAPL.US", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: "AAPL.US"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.Count)
		assert.Equal(t, "US", resp.Exchange)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "AAPL.US", resp.Data[0].Symbol)
		assert.Equal(t, "us_equity", resp.Data[0].Ext.Type)
		assert.Equal(t, int64(15400000000), resp.Data[0].Ext.FloatShares)
	})

	t.Run("multiple symbols across exchanges", func(t *testing.T) {
		symbols := "600000.SH,000001.SZ,AAPL.US,00700.HK"
		expectedResp := &GetInstrumentMetaDataResp{
			Exchange: "MIXED",
			Count:    4,
			Data: []Instrument{
				{
					Symbol:   "600000.SH",
					Exchange: "SH",
					Code:     "600000",
					Region:   "CN",
					Name:     "浦发银行",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "cn_equity",
						FloatShares: 29352000000,
						TotalShares: 29352000000,
						LimitUp:     10.45,
						LimitDown:   8.55,
						ListingDate: "1999-11-10",
						NameEn:      "Shanghai Pudong Development Bank",
						TickSize:    0.01,
					},
				},
				{
					Symbol:   "000001.SZ",
					Exchange: "SZ",
					Code:     "000001",
					Region:   "CN",
					Name:     "平安银行",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "cn_equity",
						FloatShares: 19400000000,
						TotalShares: 19400000000,
					},
				},
				{
					Symbol:   "AAPL.US",
					Exchange: "US",
					Code:     "AAPL",
					Region:   "US",
					Name:     "Apple Inc.",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "us_equity",
						FloatShares: 15400000000,
						TotalShares: 15400000000,
					},
				},
				{
					Symbol:   "00700.HK",
					Exchange: "HK",
					Code:     "00700",
					Region:   "HK",
					Name:     "Tencent Holdings Ltd",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "hk_equity",
						FloatShares: 9200000000,
						TotalShares: 9520000000,
						LotSize:     100,
					},
				},
			},
		}

		ts := setupMockServer(t, "/v1/instruments", "symbols="+symbols, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 4, resp.Count)
		assert.Len(t, resp.Data, 4)

		assert.Equal(t, "600000.SH", resp.Data[0].Symbol)
		assert.Equal(t, "cn_equity", resp.Data[0].Ext.Type)
		assert.Equal(t, 10.45, resp.Data[0].Ext.LimitUp)

		assert.Equal(t, "00700.HK", resp.Data[3].Symbol)
		assert.Equal(t, "hk_equity", resp.Data[3].Ext.Type)
		assert.Equal(t, 100, resp.Data[3].Ext.LotSize)
	})

	t.Run("etf instrument type", func(t *testing.T) {
		expectedResp := &GetInstrumentMetaDataResp{
			Exchange: "US",
			Count:    1,
			Data: []Instrument{
				{
					Symbol:   "SPY.US",
					Exchange: "US",
					Code:     "SPY",
					Region:   "US",
					Name:     "SPDR S&P 500 ETF Trust",
					Type:     InstrumentTypeETF,
					Ext:      nil,
				},
			},
		}

		ts := setupMockServer(t, "/v1/instruments", "symbols=SPY.US", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: "SPY.US"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, InstrumentTypeETF, resp.Data[0].Type)
		assert.Nil(t, resp.Data[0].Ext)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/instruments", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: INVALID",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: "INVALID"})
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
		resp, err := tf.GetInstrumentMetaData(context.Background(), &GetInstrumentMetaDataReq{Symbols: "AAPL.US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBatchGetInstrumentMetaData(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: []string{}})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("nil symbols slice returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: nil})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("too many symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		symbols := make([]string, 1001)
		for i := range symbols {
			symbols[i] = "TEST.SH"
		}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: symbols})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrTooManySymbols)
	})

	t.Run("exactly 1000 symbols is allowed", func(t *testing.T) {
		symbols := make([]string, 1000)
		for i := range symbols {
			symbols[i] = "TEST.SH"
		}
		expectedResp := &BatchGetInstrumentMetaDataResp{
			Data: []Instrument{},
		}

		ts := setupBatchMockServer(t, "/v1/instruments", symbols, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("successful batch query with mixed exchanges", func(t *testing.T) {
		symbols := []string{"600000.SH", "000001.SZ", "AAPL.US", "00700.HK"}
		expectedResp := &BatchGetInstrumentMetaDataResp{
			Data: []Instrument{
				{
					Symbol:   "600000.SH",
					Exchange: "SH",
					Code:     "600000",
					Region:   "CN",
					Name:     "浦发银行",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "cn_equity",
						FloatShares: 29352000000,
						TotalShares: 29352000000,
						LimitUp:     10.45,
						LimitDown:   8.55,
						ListingDate: "1999-11-10",
						NameEn:      "Shanghai Pudong Development Bank",
						TickSize:    0.01,
					},
				},
				{
					Symbol:   "000001.SZ",
					Exchange: "SZ",
					Code:     "000001",
					Region:   "CN",
					Name:     "平安银行",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "cn_equity",
						FloatShares: 19400000000,
						TotalShares: 19400000000,
					},
				},
				{
					Symbol:   "AAPL.US",
					Exchange: "US",
					Code:     "AAPL",
					Region:   "US",
					Name:     "Apple Inc.",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "us_equity",
						FloatShares: 15400000000,
						TotalShares: 15400000000,
					},
				},
				{
					Symbol:   "00700.HK",
					Exchange: "HK",
					Code:     "00700",
					Region:   "HK",
					Name:     "Tencent Holdings Ltd",
					Type:     InstrumentTypeStock,
					Ext: &InstrumentExt{
						Type:        "hk_equity",
						FloatShares: 9200000000,
						TotalShares: 9520000000,
						LotSize:     100,
					},
				},
			},
		}

		ts := setupBatchMockServer(t, "/v1/instruments", symbols, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 4)

		// 验证 A 股标的
		assert.Equal(t, "600000.SH", resp.Data[0].Symbol)
		assert.Equal(t, "cn_equity", resp.Data[0].Ext.Type)
		assert.Equal(t, int64(29352000000), resp.Data[0].Ext.FloatShares)
		assert.Equal(t, 10.45, resp.Data[0].Ext.LimitUp)

		// 验证美股标的
		assert.Equal(t, "AAPL.US", resp.Data[2].Symbol)
		assert.Equal(t, "us_equity", resp.Data[2].Ext.Type)

		// 验证港股标的
		assert.Equal(t, "00700.HK", resp.Data[3].Symbol)
		assert.Equal(t, "hk_equity", resp.Data[3].Ext.Type)
		assert.Equal(t, 100, resp.Data[3].Ext.LotSize)
	})

	t.Run("single symbol query", func(t *testing.T) {
		symbols := []string{"SPY.US"}
		expectedResp := &BatchGetInstrumentMetaDataResp{
			Data: []Instrument{
				{
					Symbol:   "SPY.US",
					Exchange: "US",
					Code:     "SPY",
					Region:   "US",
					Name:     "SPDR S&P 500 ETF Trust",
					Type:     InstrumentTypeETF,
					Ext:      nil,
				},
			},
		}

		ts := setupBatchMockServer(t, "/v1/instruments", symbols, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "SPY.US", resp.Data[0].Symbol)
		assert.Equal(t, InstrumentTypeETF, resp.Data[0].Type)
		assert.Nil(t, resp.Data[0].Ext)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/instruments", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: INVALID",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: []string{"INVALID"}})
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
		resp, err := tf.BatchGetInstrumentMetaData(context.Background(), &BatchGetInstrumentMetaDataReq{Symbols: []string{"AAPL.US"}})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// setupBatchMockServer 创建模拟 HTTP POST 服务器，验证请求体中的 symbols
func setupBatchMockServer(t *testing.T, expectedPath string, expectedSymbols []string, respBody any) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, http.MethodPost, r.Method)

		// 验证请求路径
		assert.Equal(t, expectedPath, r.URL.Path)

		// 验证 x-api-key header
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

		// 验证 Content-Type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 验证请求体中的 symbols
		var reqBody BatchGetInstrumentMetaDataReq
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, expectedSymbols, reqBody.Symbols)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(respBody)
		assert.Nil(t, err)
	}))
}

// setupMockServer 创建模拟 HTTP 服务器，验证请求方法和路径
func setupMockServer(t *testing.T, expectedPath, expectedQuery string, respBody any) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, http.MethodGet, r.Method)

		// 验证请求路径
		assert.Equal(t, expectedPath, r.URL.Path)

		// 验证 x-api-key header
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

		// 验证查询参数
		if expectedQuery != "" {
			assert.Equal(t, expectedQuery, r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(respBody)
		assert.Nil(t, err)
	}))
}
