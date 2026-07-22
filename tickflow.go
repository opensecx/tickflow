// Package tickflow is a Go client library for the TickFlow financial data API
// (https://api.tickflow.org), providing access to stocks, ETFs, indices, bonds,
// funds, and options across US, CN, and HK markets.
//
// All API endpoints are authenticated via the x-api-key header.
//
// Quick start:
//
//	client, err := tickflow.NewTickFlow("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	quotes, err := client.GetQuote(ctx, &tickflow.GetQuoteReq{
//	    Symbols: "AAPL.US,600000.SH",
//	})
package tickflow

import (
	"errors"
)

const (
	// defaultBaseURL is the default base URL for the TickFlow API.
	defaultBaseURL = "https://api.tickflow.org"
	// MaxBatchSymbols is the maximum number of symbols allowed in a single batch request.
	MaxBatchSymbols = 1000
)

var (
	// ErrNilReq is returned when a nil request struct is passed to an API method.
	ErrNilReq = errors.New("nil req")
	// ErrEmptyKey is returned when an empty API key is provided to NewTickFlow.
	ErrEmptyKey = errors.New("empty key")
	// ErrInvalidExchange is returned when an exchange code is not one of the supported values.
	ErrInvalidExchange = errors.New("invalid exchange, must be one of: US, SH, SZ, BJ, HK")
	// ErrEmptyID is returned when an empty ID is passed to an API method that requires one.
	ErrEmptyID = errors.New("id must not be empty")
	// ErrEmptySymbol is returned when an empty symbol is passed to an API method that requires one.
	ErrEmptySymbol = errors.New("symbol must not be empty")
	// ErrEmptySymbols is returned when an empty symbols list is passed to an API method that requires at least one.
	ErrEmptySymbols = errors.New("symbols must not be empty")
	// ErrTooManySymbols is returned when the number of symbols exceeds MaxBatchSymbols.
	ErrTooManySymbols = errors.New("symbols must not exceed 1000")
)

// validExchanges is the set of supported exchange codes.
var validExchanges = map[string]bool{
	"US": true,
	"SH": true,
	"SZ": true,
	"BJ": true,
	"HK": true,
}

// TickFlow is a client for the TickFlow financial data API.
type TickFlow struct {
	xApiKey string // http 鉴权key
	baseURL string // API 基础地址
}

// NewTickFlow creates a new TickFlow client with the given API key.
// The key is used to authenticate requests via the x-api-key header.
func NewTickFlow(key string) (*TickFlow, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}

	return &TickFlow{
		xApiKey: key,
		baseURL: defaultBaseURL,
	}, nil
}
