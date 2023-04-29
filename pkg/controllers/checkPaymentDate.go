package controllers

import (
	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/KasperKornak/StockApp/pkg/models"
	"github.com/piquette/finance-go/equity"
)

func GetPaymentDate() {
	Client := config.MongoConnect()
	stockSlice := models.ModelGetStocks(Client)
	var tickers []string

	for _, company := range stockSlice {
		tickers = append(tickers, company.Ticker)
	}

	iter := equity.List(tickers)
	for iter.Next() {
		q := iter.Equity()

	}

}
