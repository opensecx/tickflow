package tickflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetQuote(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetQuote(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("successful query with symbols", func(t *testing.T) {
		expectedResp := &GetQuoteResp{
			Data: []Quote{
				{
					Symbol:    "AAPL.US",
					Timestamp: 1704067200000,
					LastPrice: 185.5,
					Open:      184.0,
					High:      186.0,
					Low:       183.5,
					PrevClose: 183.0,
					Volume:    50000000,
					Amount:    9250000000,
					Region:    RegionUS,
					Session:   SessionRegular,
					Ext:       nil,
				},
				{
					Symbol:    "600000.SH",
					Timestamp: 1704067200000,
					LastPrice: 10.5,
					Open:      10.3,
					High:      10.6,
					Low:       10.2,
					PrevClose: 10.2,
					Volume:    100000000,
					Amount:    1050000000,
					Region:    RegionCN,
					Session:   SessionRegular,
					Ext: &QuoteExtension{
						Type: "cn_equity",
					},
				},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/quotes", map[string]string{"symbols": "AAPL.US,600000.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetQuote(&GetQuoteReq{Symbols: "AAPL.US,600000.SH"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Data, 2)

		// 验证美股
		assert.Equal(t, "AAPL.US", resp.Data[0].Symbol)
		assert.Equal(t, 185.5, resp.Data[0].LastPrice)
		assert.Equal(t, RegionUS, resp.Data[0].Region)
		assert.Equal(t, SessionRegular, resp.Data[0].Session)
		assert.Nil(t, resp.Data[0].Ext)

		// 验证 A 股
		assert.Equal(t, "600000.SH", resp.Data[1].Symbol)
		assert.Equal(t, 10.5, resp.Data[1].LastPrice)
		assert.Equal(t, RegionCN, resp.Data[1].Region)
		require.NotNil(t, resp.Data[1].Ext)
		assert.Equal(t, "cn_equity", resp.Data[1].Ext.Type)
	})

	t.Run("successful query with universes", func(t *testing.T) {
		expectedResp := &GetQuoteResp{
			Data: []Quote{
				{
					Symbol:    "00700.HK",
					Timestamp: 1704067200000,
					LastPrice: 380.0,
					Region:    RegionHK,
					Session:   SessionRegular,
				},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/quotes", map[string]string{"universes": "CN_Equity_A,CN_ETF"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetQuote(&GetQuoteReq{Universes: "CN_Equity_A,CN_ETF"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 1)
	})

	t.Run("empty request queries all", func(t *testing.T) {
		expectedResp := &GetQuoteResp{
			Data: []Quote{},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/quotes", map[string]string{}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetQuote(&GetQuoteReq{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Data)
	})

	t.Run("trading session statuses", func(t *testing.T) {
		expectedResp := &GetQuoteResp{
			Data: []Quote{
				{Symbol: "AAPL.US", Session: SessionPreMarket},
				{Symbol: "TSLA.US", Session: SessionAfterHours},
				{Symbol: "600000.SH", Session: SessionLunchBreak},
				{Symbol: "00700.HK", Session: SessionClosed},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/quotes", map[string]string{"symbols": "AAPL.US,TSLA.US,600000.SH,00700.HK"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetQuote(&GetQuoteReq{Symbols: "AAPL.US,TSLA.US,600000.SH,00700.HK"})
		require.NoError(t, err)
		require.Len(t, resp.Data, 4)
		assert.Equal(t, SessionPreMarket, resp.Data[0].Session)
		assert.Equal(t, SessionAfterHours, resp.Data[1].Session)
		assert.Equal(t, SessionLunchBreak, resp.Data[2].Session)
		assert.Equal(t, SessionClosed, resp.Data[3].Session)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/quotes", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetQuote(&GetQuoteReq{Symbols: "BAD"})
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
		resp, err := tf.GetQuote(&GetQuoteReq{Symbols: "AAPL.US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBatchGetQuote(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetQuote(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("successful batch query", func(t *testing.T) {
		symbols := []string{"AAPL.US", "600000.SH", "00700.HK"}
		expectedResp := &BatchGetQuoteResp{
			Data: []Quote{
				{
					Symbol:    "AAPL.US",
					Timestamp: 1704067200000,
					LastPrice: 185.5,
					Open:      184.0,
					High:      186.0,
					Low:       183.5,
					PrevClose: 183.0,
					Volume:    50000000,
					Amount:    9250000000,
					Region:    RegionUS,
					Session:   SessionRegular,
				},
				{
					Symbol:    "600000.SH",
					Timestamp: 1704067200000,
					LastPrice: 10.5,
					Region:    RegionCN,
					Session:   SessionRegular,
					Ext:       &QuoteExtension{Type: "cn_equity"},
				},
				{
					Symbol:    "00700.HK",
					Timestamp: 1704067200000,
					LastPrice: 380.0,
					Region:    RegionHK,
					Session:   SessionRegular,
					Ext:       &QuoteExtension{Type: "hk_equity"},
				},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodPost, "/v1/quotes", map[string]string{}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetQuote(&BatchGetQuoteReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Data, 3)

		assert.Equal(t, "AAPL.US", resp.Data[0].Symbol)
		assert.Equal(t, 185.5, resp.Data[0].LastPrice)

		assert.Equal(t, "600000.SH", resp.Data[1].Symbol)
		require.NotNil(t, resp.Data[1].Ext)
		assert.Equal(t, "cn_equity", resp.Data[1].Ext.Type)

		assert.Equal(t, "00700.HK", resp.Data[2].Symbol)
		require.NotNil(t, resp.Data[2].Ext)
		assert.Equal(t, "hk_equity", resp.Data[2].Ext.Type)
	})

	t.Run("successful batch query with mixed exchanges", func(t *testing.T) {
		symbols := []string{"AAPL.US", "TSLA.US", "600519.SH", "000001.SZ", "00700.HK"}
		expectedResp := &BatchGetQuoteResp{
			Data: []Quote{
				{Symbol: "AAPL.US", LastPrice: 185.5, Region: RegionUS},
				{Symbol: "TSLA.US", LastPrice: 250.0, Region: RegionUS},
				{Symbol: "600519.SH", LastPrice: 1700.0, Region: RegionCN},
				{Symbol: "000001.SZ", LastPrice: 12.5, Region: RegionCN},
				{Symbol: "00700.HK", LastPrice: 380.0, Region: RegionHK},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodPost, "/v1/quotes", map[string]string{}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetQuote(&BatchGetQuoteReq{Symbols: symbols})
		require.NoError(t, err)
		require.Len(t, resp.Data, 5)

		assert.Equal(t, RegionUS, resp.Data[0].Region)
		assert.Equal(t, RegionCN, resp.Data[2].Region)
		assert.Equal(t, RegionHK, resp.Data[4].Region)
	})

	t.Run("empty symbols list", func(t *testing.T) {
		expectedResp := &BatchGetQuoteResp{
			Data: []Quote{},
		}

		ts := setupQuoteMockServer(t, http.MethodPost, "/v1/quotes", map[string]string{}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetQuote(&BatchGetQuoteReq{Symbols: []string{}})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Data)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/quotes", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetQuote(&BatchGetQuoteReq{Symbols: []string{"BAD"}})
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
		resp, err := tf.BatchGetQuote(&BatchGetQuoteReq{Symbols: []string{"AAPL.US"}})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetDepth(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetDepth(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbol returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetDepth(&GetDepthReq{Symbol: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbol)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetDepthResp{
			Data: &MarketDepth{
				Symbol:     "600000.SH",
				Timestamp:  1704067200000,
				BidPrices:  []float64{10.45, 10.44, 10.43, 10.42, 10.41},
				BidVolumes: []int64{50000, 80000, 120000, 60000, 90000},
				AskPrices:  []float64{10.46, 10.47, 10.48, 10.49, 10.50},
				AskVolumes: []int64{70000, 100000, 150000, 80000, 110000},
				Region:     RegionCN,
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/depth", map[string]string{"symbol": "600000.SH"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetDepth(&GetDepthReq{Symbol: "600000.SH"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)

		depth := resp.Data
		assert.Equal(t, "600000.SH", depth.Symbol)
		assert.Equal(t, int64(1704067200000), depth.Timestamp)
		assert.Equal(t, RegionCN, depth.Region)

		// 验证五档买入（降序）
		assert.Equal(t, []float64{10.45, 10.44, 10.43, 10.42, 10.41}, depth.BidPrices)
		assert.Equal(t, []int64{50000, 80000, 120000, 60000, 90000}, depth.BidVolumes)

		// 验证五档卖出（升序）
		assert.Equal(t, []float64{10.46, 10.47, 10.48, 10.49, 10.50}, depth.AskPrices)
		assert.Equal(t, []int64{70000, 100000, 150000, 80000, 110000}, depth.AskVolumes)
	})

	t.Run("successful query for US stock", func(t *testing.T) {
		expectedResp := &GetDepthResp{
			Data: &MarketDepth{
				Symbol:     "AAPL.US",
				Timestamp:  1704067200000,
				BidPrices:  []float64{185.40, 185.35, 185.30, 185.25, 185.20},
				BidVolumes: []int64{500, 800, 1200, 600, 900},
				AskPrices:  []float64{185.45, 185.50, 185.55, 185.60, 185.65},
				AskVolumes: []int64{700, 1000, 1500, 800, 1100},
				Region:     RegionUS,
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/depth", map[string]string{"symbol": "AAPL.US"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetDepth(&GetDepthReq{Symbol: "AAPL.US"})
		require.NoError(t, err)
		require.NotNil(t, resp.Data)
		assert.Equal(t, "AAPL.US", resp.Data.Symbol)
		assert.Equal(t, RegionUS, resp.Data.Region)
		assert.Len(t, resp.Data.BidPrices, 5)
		assert.Len(t, resp.Data.AskPrices, 5)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/depth", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "SYMBOL_NOT_FOUND",
				Message: "Symbol not found: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetDepth(&GetDepthReq{Symbol: "BAD"})
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
		resp, err := tf.GetDepth(&GetDepthReq{Symbol: "600000.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBatchGetDepth(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetDepth(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetDepth(&BatchGetDepthReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful batch query", func(t *testing.T) {
		symbols := "600000.SH,AAPL.US"
		expectedResp := &BatchGetDepthResp{
			Data: map[string]*MarketDepth{
				"600000.SH": {
					Symbol:     "600000.SH",
					Timestamp:  1704067200000,
					BidPrices:  []float64{10.45, 10.44, 10.43, 10.42, 10.41},
					BidVolumes: []int64{50000, 80000, 120000, 60000, 90000},
					AskPrices:  []float64{10.46, 10.47, 10.48, 10.49, 10.50},
					AskVolumes: []int64{70000, 100000, 150000, 80000, 110000},
					Region:     RegionCN,
				},
				"AAPL.US": {
					Symbol:     "AAPL.US",
					Timestamp:  1704067200000,
					BidPrices:  []float64{185.40, 185.35, 185.30, 185.25, 185.20},
					BidVolumes: []int64{500, 800, 1200, 600, 900},
					AskPrices:  []float64{185.45, 185.50, 185.55, 185.60, 185.65},
					AskVolumes: []int64{700, 1000, 1500, 800, 1100},
					Region:     RegionUS,
				},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/depth/batch", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetDepth(&BatchGetDepthReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 2)

		// 验证 A 股
		sh := resp.Data["600000.SH"]
		require.NotNil(t, sh)
		assert.Equal(t, "600000.SH", sh.Symbol)
		assert.Equal(t, RegionCN, sh.Region)
		assert.Equal(t, []float64{10.45, 10.44, 10.43, 10.42, 10.41}, sh.BidPrices)
		assert.Equal(t, []float64{10.46, 10.47, 10.48, 10.49, 10.50}, sh.AskPrices)

		// 验证美股
		us := resp.Data["AAPL.US"]
		require.NotNil(t, us)
		assert.Equal(t, "AAPL.US", us.Symbol)
		assert.Equal(t, RegionUS, us.Region)
		assert.Equal(t, []float64{185.40, 185.35, 185.30, 185.25, 185.20}, us.BidPrices)
		assert.Equal(t, []float64{185.45, 185.50, 185.55, 185.60, 185.65}, us.AskPrices)
	})

	t.Run("successful batch query with mixed exchanges", func(t *testing.T) {
		symbols := "600519.SH,00700.HK,AAPL.US"
		expectedResp := &BatchGetDepthResp{
			Data: map[string]*MarketDepth{
				"600519.SH": {
					Symbol:    "600519.SH",
					BidPrices: []float64{1700.0, 1699.5, 1699.0, 1698.5, 1698.0},
					AskPrices: []float64{1700.5, 1701.0, 1701.5, 1702.0, 1702.5},
					Region:    RegionCN,
				},
				"00700.HK": {
					Symbol:    "00700.HK",
					BidPrices: []float64{379.5, 379.0, 378.5, 378.0, 377.5},
					AskPrices: []float64{380.0, 380.5, 381.0, 381.5, 382.0},
					Region:    RegionHK,
				},
				"AAPL.US": {
					Symbol:    "AAPL.US",
					BidPrices: []float64{185.40, 185.35, 185.30, 185.25, 185.20},
					AskPrices: []float64{185.45, 185.50, 185.55, 185.60, 185.65},
					Region:    RegionUS,
				},
			},
		}

		ts := setupQuoteMockServer(t, http.MethodGet, "/v1/depth/batch", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetDepth(&BatchGetDepthReq{Symbols: symbols})
		require.NoError(t, err)
		require.Len(t, resp.Data, 3)
		assert.Equal(t, RegionCN, resp.Data["600519.SH"].Region)
		assert.Equal(t, RegionHK, resp.Data["00700.HK"].Region)
		assert.Equal(t, RegionUS, resp.Data["AAPL.US"].Region)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/depth/batch", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BAD",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetDepth(&BatchGetDepthReq{Symbols: "BAD"})
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
		resp, err := tf.BatchGetDepth(&BatchGetDepthReq{Symbols: "600000.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// setupQuoteMockServer 创建模拟 HTTP 行情服务器，验证请求方法、路径和查询参数
// expectedParams 为期望的查询参数键值对（无需考虑顺序，自动处理 URL 编码）
func setupQuoteMockServer(t *testing.T, expectedMethod, expectedPath string, expectedParams map[string]string, respBody any) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		assert.Equal(t, expectedMethod, r.Method)

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
