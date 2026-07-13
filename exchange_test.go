package tickflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExchangeInstrument(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}
		resp, err := tf.GetExchangeInstrument(nil)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrNilReq)
	})

	t.Run("invalid exchange returns error", func(t *testing.T) {
		tf := &TickFlow{xApiKey: "test-key", baseURL: defaultBaseURL}

		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "XX"})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidExchange)

		resp, err = tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "us"})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidExchange)

		resp, err = tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: ""})
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrInvalidExchange)
	})

	t.Run("successful query without type filter", func(t *testing.T) {
		expectedResp := &GetExchangeInstrumentResp{
			Exchange: "US",
			Count:    2,
			Data: []Instrument{
				{
					Symbol:   "AAPL.US",
					Exchange: "US",
					Code:     "AAPL",
					Region:   "US",
					Name:     strPtr("Apple Inc."),
					Type:     typePtr(InstrumentTypeStock),
					Ext: &InstrumentExt{
						Type:        "us_equity",
						FloatShares: 15400000000,
						TotalShares: 15400000000,
					},
				},
				{
					Symbol:   "MSFT.US",
					Exchange: "US",
					Code:     "MSFT",
					Region:   "US",
					Name:     strPtr("Microsoft Corporation"),
					Type:     typePtr(InstrumentTypeStock),
					Ext: &InstrumentExt{
						Type:        "us_equity",
						FloatShares: 7420000000,
						TotalShares: 7420000000,
					},
				},
			},
		}

		ts := setupMockServer(t, "/v1/exchanges/US/instruments", "", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "US"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "US", resp.Exchange)
		assert.Equal(t, 2, resp.Count)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "AAPL.US", resp.Data[0].Symbol)
		assert.Equal(t, "MSFT.US", resp.Data[1].Symbol)
	})

	t.Run("successful query with type filter", func(t *testing.T) {
		expectedResp := &GetExchangeInstrumentResp{
			Exchange: "US",
			Count:    1,
			Data: []Instrument{
				{
					Symbol:   "SPY.US",
					Exchange: "US",
					Code:     "SPY",
					Region:   "US",
					Name:     strPtr("SPDR S&P 500 ETF Trust"),
					Type:     typePtr(InstrumentTypeETF),
					Ext:      nil,
				},
			},
		}

		ts := setupMockServer(t, "/v1/exchanges/US/instruments", "type=etf", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "US", Type: "etf"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.Count)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, InstrumentTypeETF, *resp.Data[0].Type)
		assert.Nil(t, resp.Data[0].Ext)
	})

	t.Run("successful query for HK exchange", func(t *testing.T) {
		expectedResp := &GetExchangeInstrumentResp{
			Exchange: "HK",
			Count:    1,
			Data: []Instrument{
				{
					Symbol:   "00700.HK",
					Exchange: "HK",
					Code:     "00700",
					Region:   "HK",
					Name:     strPtr("Tencent Holdings Ltd"),
					Type:     typePtr(InstrumentTypeStock),
					Ext: &InstrumentExt{
						Type:        "hk_equity",
						FloatShares: 9200000000,
						TotalShares: 9520000000,
						LotSize:     100,
					},
				},
			},
		}

		ts := setupMockServer(t, "/v1/exchanges/HK/instruments", "", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "HK"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "HK", resp.Exchange)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "00700.HK", resp.Data[0].Symbol)
		assert.Equal(t, "hk_equity", resp.Data[0].Ext.Type)
		assert.Equal(t, 100, resp.Data[0].Ext.LotSize)
	})

	t.Run("successful query for SH exchange with cn_equity", func(t *testing.T) {
		expectedResp := &GetExchangeInstrumentResp{
			Exchange: "SH",
			Count:    1,
			Data: []Instrument{
				{
					Symbol:   "600000.SH",
					Exchange: "SH",
					Code:     "600000",
					Region:   "CN",
					Name:     strPtr("浦发银行"),
					Type:     typePtr(InstrumentTypeStock),
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
			},
		}

		ts := setupMockServer(t, "/v1/exchanges/SH/instruments", "", expectedResp)
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "SH"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "SH", resp.Exchange)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "600000.SH", resp.Data[0].Symbol)
		assert.Equal(t, "cn_equity", resp.Data[0].Ext.Type)
		assert.Equal(t, 10.45, resp.Data[0].Ext.LimitUp)
		assert.Equal(t, "1999-11-10", resp.Data[0].Ext.ListingDate)
	})

	t.Run("server error response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/exchanges/US/instruments", r.URL.Path)
			assert.Equal(t, "test-key", r.Header.Get("x-api-key"))

			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ApiError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			})
			assert.Nil(t, err)
		}))
		defer ts.Close()

		tf := &TickFlow{xApiKey: "test-key", baseURL: ts.URL}
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "US"})
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
		resp, err := tf.GetExchangeInstrument(&GetExchangeInstrumentReq{Exchange: "US"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
