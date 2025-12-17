package kline

import "strconv"

type KlineEvent struct {
	Event  string `json:"e"`
	Time   int64  `json:"E"`
	Symbol string `json:"s"`
	Kline  Kline  `json:"k"`
}

type Kline struct {
	StartTime    int64  `json:"t"`
	CloseTime    int64  `json:"T"`
	Symbol       string `json:"s"`
	Interval     string `json:"i"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"L"`
	Open         string `json:"o"`
	Close        string `json:"c"`
	High         string `json:"h"`
	Low          string `json:"l"`
	Volume       string `json:"v"`
	Trades       int64  `json:"n"`
	IsClosed     bool   `json:"x"`
	QuoteVolume  string `json:"q"`
	TakerBuyBase string `json:"V"`
	TakerBuyQuote string `json:"Q"`
}

func (k *Kline) GetClosePrice() (float64, error) {
	return strconv.ParseFloat(k.Close, 64)
}
