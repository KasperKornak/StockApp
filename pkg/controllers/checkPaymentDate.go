package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/KasperKornak/StockApp/pkg/models"
	"github.com/piquette/finance-go/equity"
	"github.com/piquette/finance-go/forex"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
		yahooDate := q.DividendDate
		for _, company := range stockSlice {
			if (q.Symbol == company.Ticker) && (q.DividendDate > company.NextPayment) {
				next := company.NextPayment
				filter := bson.M{"ticker": company.Ticker}
				stocks := Client.Database("stock").Collection("tickers")
				update := bson.M{
					"$set": bson.M{
						"nextpayment": yahooDate,
						"prevpayment": next,
					},
				}
				_, err := stocks.UpdateOne(context.TODO(), filter, update)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Updated dividend for: %s", company.Ticker)
			}
		}
	}
	err := Client.Disconnect(context.TODO())
	if err != nil {
		panic(err)
	}
}

func CheckPayment() {
	Client := config.MongoConnect()
	stockSlice := models.ModelGetStocks(Client)
	for _, company := range stockSlice {
		if (company.NextPayment <= int(time.Now().Unix())) && (company.NextPayment != company.PrevPayment) {
			pair := fmt.Sprintf("%sPLN=x", company.Currency)
			q, err := forex.Get(pair)
			if err != nil {
				fmt.Println(err)
			}

			noShares := company.Shares
			div := company.DivQuarterlyRate
			divPLNtoSend := div * float64(noShares) * q.Bid * float64(company.Domestictax) / 100.0

			var divUSDtoSend float64
			if company.Currency != "USD" {
				pairCorr := fmt.Sprintf("%sUSD=x", company.Currency)
				correction, err := forex.Get(pairCorr)
				if err != nil {
					fmt.Println(err)
				}

				divUSDtoSend = div * float64(noShares) * correction.Bid
			} else {
				divUSDtoSend = div * float64(noShares)
			}
			filter := bson.M{"ticker": company.Ticker}
			stocks := Client.Database("stock").Collection("tickers")
			update := bson.M{"$set": bson.M{"divytd": (company.DivYTD + divUSDtoSend)}}
			_, err = stocks.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				panic(err)
			}

			updated := bson.M{"$set": bson.M{"divpln": (company.DivPLN + divPLNtoSend)}}
			_, err = stocks.UpdateOne(context.TODO(), filter, updated)
			if err != nil {
				panic(err)
			}

			updateNextDate := bson.M{"$set": bson.M{"nextpayment:": (company.PrevPayment)}}
			_, err = stocks.UpdateOne(context.TODO(), filter, updateNextDate)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Recieved dividend from: %s\n", company.Ticker)
			fmt.Printf("Amount in USD: %f\n", divUSDtoSend)
			fmt.Printf("Tax to pay in PLN: %f\n", divPLNtoSend)
		}
	}
	err := Client.Disconnect(context.TODO())
	if err != nil {
		panic(err)
	}
}

func UpdateSummary() {
	var updatedDoc models.DeletedCompany
	updatedDoc.Year = time.Now().Year()
	updatedDoc.Ticker = "YEAR_SUMMARY"
	updatedDoc.DivYTD = 0.0
	updatedDoc.DivPLN = 0.0

	Client := config.MongoConnect()
	tickers := Client.Database("stock").Collection("tickers")
	stockSlice := models.ModelGetStocks(Client)
	divTax := 0.0
	divRec := 0.0

	for _, stock := range stockSlice {
		divTax = divTax + stock.DivPLN
		divRec = divRec + stock.DivYTD
	}

	deletedStocks := models.ModelGetStockByTicker("DELETED_SUM", Client)

	divTax = divTax + deletedStocks.DivPLN
	divRec = divRec + deletedStocks.DivYTD

	updatedDoc.DivPLN = divTax
	updatedDoc.DivYTD = divRec

	filter := bson.M{"ticker": "YEAR_SUMMARY"}
	update := bson.M{"$set": updatedDoc}
	_, err := tickers.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		panic(err)
	}
	err = Client.Disconnect(context.TODO())
	if err != nil {
		panic(err)
	}
}

func CheckYear() {
	Client := config.MongoConnect()
	tickers := Client.Database("stock").Collection("tickers")
	defer Client.Disconnect(context.TODO())
	err := tickers.FindOne(context.TODO(), bson.M{"ticker": "DELETED_SUM", "year": time.Now().Year()}).Err()

	if err == mongo.ErrNoDocuments {
		var deleted models.DeletedCompany
		deleted.Year = time.Now().Year()
		deleted.Ticker = "DELETED_SUM"
		deleted.DivYTD = 0.0
		deleted.DivPLN = 0.0

		_, err := tickers.InsertOne(context.TODO(), &deleted)
		if err != nil {
			panic(err)
		}

		var newDocument models.DeletedCompany
		newDocument.Year = time.Now().Year()
		newDocument.Ticker = "YEAR_SUMMARY"
		newDocument.DivYTD = 0.0
		newDocument.DivPLN = 0.0

		_, err = tickers.InsertOne(context.TODO(), &newDocument)
		if err != nil {
			panic(err)
		}

		stockSlice := models.ModelGetStocks(Client)
		for _, stock := range stockSlice {
			stock.DivPLN = 0.0
			stock.DivYTD = 0.0
			filter := bson.M{"ticker": stock.Ticker}
			update := bson.M{"$set": stock}
			_, err := tickers.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("Happy New Year!")
	}
}
