package models

type Company struct {
	Ticker      string `bson:"ticker"`
	Shares      int    `bson:"shares"`
	DomesticTax int    `bson:"domestictax"`
}
