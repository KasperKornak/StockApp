package models

type Company struct {
	Ticker      string `json:"ticker"`
	Shares      int    `json:"shares"`
	DomesticTax int    `json:"domestictax"`
}
