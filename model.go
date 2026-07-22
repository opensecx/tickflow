package tickflow

// InstrumentType represents the type of a financial instrument.
type InstrumentType string

const (
	// InstrumentTypeStock represents a stock (equity).
	InstrumentTypeStock InstrumentType = "stock"
	// InstrumentTypeETF represents an exchange-traded fund.
	InstrumentTypeETF InstrumentType = "etf"
	// InstrumentTypeIndex represents a market index.
	InstrumentTypeIndex InstrumentType = "index"
	// InstrumentTypeBond represents a bond.
	InstrumentTypeBond InstrumentType = "bond"
	// InstrumentTypeFund represents a mutual fund.
	InstrumentTypeFund InstrumentType = "fund"
	// InstrumentTypeOptions represents options contracts.
	InstrumentTypeOptions InstrumentType = "options"
	// InstrumentTypeOther represents an instrument that does not fit the other categories.
	InstrumentTypeOther InstrumentType = "other"
)

// InstrumentExt contains market-specific extension fields for an instrument.
type InstrumentExt struct {
	Type        string  `json:"type,omitempty"`         // cn_equity / us_equity / hk_equity
	ListingDate string  `json:"listing_date,omitempty"` // 上市日期 (cn_equity)
	FloatShares float64 `json:"float_shares,omitempty"` // 流通股本
	TotalShares float64 `json:"total_shares,omitempty"` // 总股本
	TickSize    float64 `json:"tick_size,omitempty"`    // 最小变动价位 (cn_equity)
	LimitUp     float64 `json:"limit_up,omitempty"`     // 涨停价 (cn_equity)
	LimitDown   float64 `json:"limit_down,omitempty"`   // 跌停价 (cn_equity)
	NameEn      string  `json:"name_en,omitempty"`      // 英文名 (cn_equity)
	LotSize     int     `json:"lot_size,omitempty"`     // 每手股数 (hk_equity)
}

// Instrument represents a financial instrument listed on an exchange.
type Instrument struct {
	Symbol   string         `json:"symbol"`   // e.g. "600000.SH", "AAPL.US"
	Exchange string         `json:"exchange"` // e.g. "SH", "US"
	Code     string         `json:"code"`     // 交易所原生代码 e.g. "600000", "AAPL"
	Region   string         `json:"region"`   // 地区: CN, US, HK
	Name     string         `json:"name"`     // 可读名称
	Type     InstrumentType `json:"type"`     // 标的类型
	Ext      *InstrumentExt `json:"ext"`      // 市场特定扩展字段
}
