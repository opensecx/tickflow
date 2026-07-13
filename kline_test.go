package tickflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// int64Ptr 返回 int64 指针
func int64Ptr(i int64) *int64 {
	return &i
}

func TestGetKline(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetKline(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbol returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbol)
	})

	t.Run("successful query with symbol only", func(t *testing.T) {
		expectedResp := &GetKlineResp{
			Data: &CompactKlineData{
				Timestamp: []int64{1704067200000, 1704153600000},
				Open:      []float64{150.0, 151.0},
				High:      []float64{155.0, 156.0},
				Low:       []float64{149.0, 150.0},
				Close:     []float64{154.0, 155.5},
				Volume:    []int64{1000000, 1200000},
				Amount:    []float64{154000000, 186600000},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines", map[string]string{"symbol": "AAPL.US"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "AAPL.US"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Equal(t, []int64{1704067200000, 1704153600000}, resp.Data.Timestamp)
		assert.Equal(t, []float64{150.0, 151.0}, resp.Data.Open)
		assert.Equal(t, []float64{154.0, 155.5}, resp.Data.Close)
		assert.Equal(t, []int64{1000000, 1200000}, resp.Data.Volume)
	})

	t.Run("successful query with all params", func(t *testing.T) {
		expectedResp := &GetKlineResp{
			Data: &CompactKlineData{
				Timestamp: []int64{1704067200000, 1704153600000, 1704240000000},
				Open:      []float64{10.0, 10.1, 10.2},
				High:      []float64{10.5, 10.6, 10.7},
				Low:       []float64{9.8, 9.9, 10.0},
				Close:     []float64{10.3, 10.4, 10.5},
				Volume:    []int64{5000000, 6000000, 7000000},
				Amount:    []float64{51500000, 62400000, 73500000},
				PrevClose: []float64{9.9, 10.0, 10.1},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines", map[string]string{
			"symbol":     "600000.SH",
			"period":     "1d",
			"count":      "3",
			"start_time": "1704067200000",
			"end_time":   "1704326400000",
			"adjust":     "forward",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{
			Symbol:    "600000.SH",
			Period:    Period1d,
			Count:     3,
			StartTime: int64Ptr(1704067200000),
			EndTime:   int64Ptr(1704326400000),
			Adjust:    AdjustTypeForward,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data.Timestamp, 3)
		assert.Equal(t, []float64{10.0, 10.1, 10.2}, resp.Data.Open)
		assert.Equal(t, []float64{9.9, 10.0, 10.1}, resp.Data.PrevClose)
	})

	t.Run("successful query with period 1m", func(t *testing.T) {
		expectedResp := &GetKlineResp{
			Data: &CompactKlineData{
				Timestamp: []int64{1704067200000, 1704067260000, 1704067320000},
				Open:      []float64{10.0, 10.05, 10.1},
				High:      []float64{10.1, 10.15, 10.2},
				Low:       []float64{9.95, 10.0, 10.05},
				Close:     []float64{10.05, 10.1, 10.15},
				Volume:    []int64{100000, 120000, 110000},
				Amount:    []float64{1005000, 1212000, 1116500},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines", map[string]string{"symbol": "00700.HK", "period": "1m"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "00700.HK", Period: Period1m})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data.Timestamp, 3)
		assert.Equal(t, Period1m, Period("1m"))
	})

	t.Run("response with optional fields", func(t *testing.T) {
		expectedResp := &GetKlineResp{
			Data: &CompactKlineData{
				Timestamp:       []int64{1704067200000},
				Open:            []float64{3000.0},
				High:            []float64{3100.0},
				Low:             []float64{2950.0},
				Close:           []float64{3050.0},
				Volume:          []int64{500000},
				Amount:          []float64{1525000000},
				PrevClose:       []float64{2980.0},
				OpenInterest:    []float64{15000.0},
				SettlementPrice: []float64{3045.0},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines", map[string]string{"symbol": "IF2401.CFE"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "IF2401.CFE"})
		require.NoError(t, err)
		require.NotNil(t, resp.Data)
		assert.Equal(t, []float64{2980.0}, resp.Data.PrevClose)
		assert.Equal(t, []float64{15000.0}, resp.Data.OpenInterest)
		assert.Equal(t, []float64{3045.0}, resp.Data.SettlementPrice)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/klines", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_PERIOD",
				Message: "Invalid period: 2d",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "AAPL.US", Period: "2d"})
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
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "AAPL.US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("404 symbol not found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "SYMBOL_NOT_FOUND",
				Message: "Symbol not found: INVALID.US",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetKline(&GetKlineReq{Symbol: "INVALID.US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBatchGetKline(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetKline(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetKline(&BatchGetKlineReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful batch query with multiple symbols", func(t *testing.T) {
		symbols := "600000.SH,000001.SZ,AAPL.US"
		expectedResp := &BatchGetKlineResp{
			Data: map[string]*CompactKlineData{
				"600000.SH": {
					Timestamp: []int64{1704067200000, 1704153600000},
					Open:      []float64{10.0, 10.1},
					High:      []float64{10.5, 10.6},
					Low:       []float64{9.8, 9.9},
					Close:     []float64{10.3, 10.4},
					Volume:    []int64{5000000, 6000000},
					Amount:    []float64{51500000, 62400000},
					PrevClose: []float64{9.9, 10.0},
				},
				"000001.SZ": {
					Timestamp: []int64{1704067200000, 1704153600000},
					Open:      []float64{20.0, 20.1},
					High:      []float64{20.5, 20.6},
					Low:       []float64{19.8, 19.9},
					Close:     []float64{20.3, 20.4},
					Volume:    []int64{3000000, 4000000},
					Amount:    []float64{60900000, 81600000},
				},
				"AAPL.US": {
					Timestamp: []int64{1704067200000, 1704153600000},
					Open:      []float64{150.0, 151.0},
					High:      []float64{155.0, 156.0},
					Low:       []float64{149.0, 150.0},
					Close:     []float64{154.0, 155.5},
					Volume:    []int64{1000000, 1200000},
					Amount:    []float64{154000000, 186600000},
				},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines/batch", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetKline(&BatchGetKlineReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 3)

		// 验证 A 股标的
		sh := resp.Data["600000.SH"]
		require.NotNil(t, sh)
		assert.Equal(t, []float64{10.0, 10.1}, sh.Open)
		assert.Equal(t, []float64{9.9, 10.0}, sh.PrevClose)

		// 验证深市标的
		sz := resp.Data["000001.SZ"]
		require.NotNil(t, sz)
		assert.Equal(t, []float64{20.0, 20.1}, sz.Open)

		// 验证美股标的
		us := resp.Data["AAPL.US"]
		require.NotNil(t, us)
		assert.Equal(t, []float64{150.0, 151.0}, us.Open)
		assert.Equal(t, []float64{154.0, 155.5}, us.Close)
	})

	t.Run("successful batch query with params", func(t *testing.T) {
		symbols := "600000.SH,AAPL.US"
		expectedResp := &BatchGetKlineResp{
			Data: map[string]*CompactKlineData{
				"600000.SH": {
					Timestamp: []int64{1704067200000},
					Open:      []float64{10.0},
					High:      []float64{10.5},
					Low:       []float64{9.8},
					Close:     []float64{10.3},
					Volume:    []int64{5000000},
					Amount:    []float64{51500000},
				},
				"AAPL.US": {
					Timestamp: []int64{1704067200000},
					Open:      []float64{150.0},
					High:      []float64{155.0},
					Low:       []float64{149.0},
					Close:     []float64{154.0},
					Volume:    []int64{1000000},
					Amount:    []float64{154000000},
				},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines/batch", map[string]string{
			"symbols": symbols,
			"period":  "1d",
			"count":   "1",
			"adjust":  "backward",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetKline(&BatchGetKlineReq{
			Symbols: symbols,
			Period:  Period1d,
			Count:   1,
			Adjust:  AdjustTypeBackward,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Data, 2)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/klines/batch", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BADFORMAT",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetKline(&BatchGetKlineReq{Symbols: "BADFORMAT"})
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
		resp, err := tf.BatchGetKline(&BatchGetKlineReq{Symbols: "AAPL.US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetExFactor(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetExFactor(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("empty symbols returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetExFactor(&GetExFactorReq{Symbols: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptySymbols)
	})

	t.Run("successful query with symbols only", func(t *testing.T) {
		symbols := "600519.SH,000001.SZ"
		expectedResp := &GetExFactorResp{
			Data: map[string][]ExFactorEntry{
				"600519.SH": {
					{Timestamp: 1704067200000, ExFactor: 1.05},
					{Timestamp: 1711843200000, ExFactor: 1.02},
				},
				"000001.SZ": {
					{Timestamp: 1704067200000, ExFactor: 1.10},
				},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines/ex-factors", map[string]string{"symbols": symbols}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExFactor(&GetExFactorReq{Symbols: symbols})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 2)

		// 验证 600519.SH 的除权因子
		sh := resp.Data["600519.SH"]
		require.Len(t, sh, 2)
		assert.Equal(t, int64(1704067200000), sh[0].Timestamp)
		assert.Equal(t, 1.05, sh[0].ExFactor)
		assert.Equal(t, int64(1711843200000), sh[1].Timestamp)
		assert.Equal(t, 1.02, sh[1].ExFactor)

		// 验证 000001.SZ 的除权因子
		sz := resp.Data["000001.SZ"]
		require.Len(t, sz, 1)
		assert.Equal(t, int64(1704067200000), sz[0].Timestamp)
		assert.Equal(t, 1.10, sz[0].ExFactor)
	})

	t.Run("successful query with time range", func(t *testing.T) {
		symbols := "AAPL.US"
		expectedResp := &GetExFactorResp{
			Data: map[string][]ExFactorEntry{
				"AAPL.US": {
					{Timestamp: 1704067200000, ExFactor: 1.0},
					{Timestamp: 1711843200000, ExFactor: 0.95},
				},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines/ex-factors", map[string]string{
			"symbols":    symbols,
			"start_time": "1700000000000",
			"end_time":   "1720000000000",
		}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExFactor(&GetExFactorReq{
			Symbols:   symbols,
			StartTime: int64Ptr(1700000000000),
			EndTime:   int64Ptr(1720000000000),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 1)

		us := resp.Data["AAPL.US"]
		require.Len(t, us, 2)
		assert.Equal(t, 1.0, us[0].ExFactor)
		assert.Equal(t, 0.95, us[1].ExFactor)
	})

	t.Run("single symbol query", func(t *testing.T) {
		expectedResp := &GetExFactorResp{
			Data: map[string][]ExFactorEntry{
				"00700.HK": {
					{Timestamp: 1709683200000, ExFactor: 1.08},
				},
			},
		}

		ts := setupKlineMockServer(t, "/v1/klines/ex-factors", map[string]string{"symbols": "00700.HK"}, expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExFactor(&GetExFactorReq{Symbols: "00700.HK"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Data, 1)
		hk := resp.Data["00700.HK"]
		require.Len(t, hk, 1)
		assert.Equal(t, int64(1709683200000), hk[0].Timestamp)
		assert.Equal(t, 1.08, hk[0].ExFactor)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/klines/ex-factors", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_SYMBOL",
				Message: "Invalid symbol format: BADFORMAT",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExFactor(&GetExFactorReq{Symbols: "BADFORMAT"})
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
		resp, err := tf.GetExFactor(&GetExFactorReq{Symbols: "600519.SH"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

// setupKlineMockServer 创建模拟 HTTP GET K线服务器，验证请求路径和查询参数
// expectedParams 为期望的查询参数键值对（无需考虑顺序）
func setupKlineMockServer(t *testing.T, expectedPath string, expectedParams map[string]string, respBody any) *httptest.Server {
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
