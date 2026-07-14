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

func TestGetUniverse(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetUniverseResp{
			Data: []UniverseSummary{
				{
					ID:          "CN_Equity_A",
					Name:        "A股股票",
					Category:    "Equity",
					Region:      "CN",
					SymbolCount: 4500,
					Description: "中国A股市场全部股票",
				},
				{
					ID:          "CN_ETF",
					Name:        "ETF基金",
					Category:    "ETF",
					Region:      "CN",
					SymbolCount: 800,
					Description: "中国ETF基金",
				},
				{
					ID:          "US_Equity",
					Name:        "US Stocks",
					Category:    "Equity",
					Region:      "US",
					SymbolCount: 6000,
					Description: "US Stock Market",
				},
			},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/universes", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(expectedResp)
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetUniverse(context.Background())
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Data, 3)

		// 验证第一个标的池
		assert.Equal(t, "CN_Equity_A", resp.Data[0].ID)
		assert.Equal(t, "A股股票", resp.Data[0].Name)
		assert.Equal(t, 4500, resp.Data[0].SymbolCount)

		// 验证第二个标的池
		assert.Equal(t, "CN_ETF", resp.Data[1].ID)
		assert.Equal(t, "ETF基金", resp.Data[1].Name)

		// 验证第三个标的池
		assert.Equal(t, "US_Equity", resp.Data[2].ID)
		assert.Equal(t, "US", resp.Data[2].Region)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetUniverse(context.Background())
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
		resp, err := tf.GetUniverse(context.Background())
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestBatchGetUniverse(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.BatchGetUniverse(context.Background(), nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("successful batch query", func(t *testing.T) {
		ids := []string{"CN_Equity_A", "CN_ETF"}
		expectedResp := &BatchGetUniverseResp{
			Data: map[string]*UniverseDetail{
				"CN_Equity_A": {
					UniverseSummary: UniverseSummary{
						ID:          "CN_Equity_A",
						Name:        "A股股票",
						Category:    "Equity",
						Region:      "CN",
						SymbolCount: 3,
						Description: "中国A股市场全部股票",
					},
					Symbols: []string{"600000.SH", "000001.SZ", "600519.SH"},
				},
				"CN_ETF": {
					UniverseSummary: UniverseSummary{
						ID:          "CN_ETF",
						Name:        "ETF基金",
						Category:    "ETF",
						Region:      "CN",
						SymbolCount: 2,
						Description: "中国ETF基金",
					},
					Symbols: []string{"510300.SH", "159915.SZ"},
				},
			},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/universes/batch", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(expectedResp)
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetUniverse(context.Background(), &BatchGetUniverseReq{IDs: ids})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)
		assert.Len(t, resp.Data, 2)

		// 验证 CN_Equity_A
		cn := resp.Data["CN_Equity_A"]
		require.NotNil(t, cn)
		assert.Equal(t, "CN_Equity_A", cn.ID)
		assert.Equal(t, "A股股票", cn.Name)
		assert.Equal(t, 3, cn.SymbolCount)
		assert.Equal(t, []string{"600000.SH", "000001.SZ", "600519.SH"}, cn.Symbols)

		// 验证 CN_ETF
		etf := resp.Data["CN_ETF"]
		require.NotNil(t, etf)
		assert.Equal(t, "CN_ETF", etf.ID)
		assert.Equal(t, []string{"510300.SH", "159915.SZ"}, etf.Symbols)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INVALID_ID",
				Message: "Invalid universe ID: BAD_ID",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.BatchGetUniverse(context.Background(), &BatchGetUniverseReq{IDs: []string{"BAD_ID"}})
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
		resp, err := tf.BatchGetUniverse(context.Background(), &BatchGetUniverseReq{IDs: []string{"CN_Equity_A"}})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetUniverseDetail(t *testing.T) {
	t.Run("empty id returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetUniverseDetail(context.Background(), "")
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrEmptyID)
	})

	t.Run("successful query", func(t *testing.T) {
		expectedResp := &GetUniverseDetailResp{
			Data: &UniverseDetail{
				UniverseSummary: UniverseSummary{
					ID:          "CN_Equity_A",
					Name:        "A股股票",
					Category:    "Equity",
					Region:      "CN",
					SymbolCount: 3,
					Description: "中国A股市场全部股票",
				},
				Symbols: []string{"600000.SH", "000001.SZ", "600519.SH"},
			},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/universes/CN_Equity_A", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(expectedResp)
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetUniverseDetail(context.Background(), "CN_Equity_A")
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Data)

		detail := resp.Data
		assert.Equal(t, "CN_Equity_A", detail.ID)
		assert.Equal(t, "A股股票", detail.Name)
		assert.Equal(t, "Equity", detail.Category)
		assert.Equal(t, "CN", detail.Region)
		assert.Equal(t, 3, detail.SymbolCount)
		assert.Equal(t, []string{"600000.SH", "000001.SZ", "600519.SH"}, detail.Symbols)
	})

	t.Run("successful query for US universe", func(t *testing.T) {
		expectedResp := &GetUniverseDetailResp{
			Data: &UniverseDetail{
				UniverseSummary: UniverseSummary{
					ID:          "US_Equity",
					Name:        "US Stocks",
					Category:    "Equity",
					Region:      "US",
					SymbolCount: 2,
					Description: "US Stock Market",
				},
				Symbols: []string{"AAPL.US", "MSFT.US"},
			},
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/universes/US_Equity", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(expectedResp)
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetUniverseDetail(context.Background(), "US_Equity")
		require.NoError(t, err)
		require.NotNil(t, resp.Data)
		assert.Equal(t, "US_Equity", resp.Data.ID)
		assert.Equal(t, "US", resp.Data.Region)
		assert.Equal(t, []string{"AAPL.US", "MSFT.US"}, resp.Data.Symbols)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "UNIVERSE_NOT_FOUND",
				Message: "Universe not found: UNKNOWN",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetUniverseDetail(context.Background(), "UNKNOWN")
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
		resp, err := tf.GetUniverseDetail(context.Background(), "CN_Equity_A")
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
