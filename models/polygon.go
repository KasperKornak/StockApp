package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
)

type PolygonJson struct {
	NextURL string `json:"next_url"`
	Results []struct {
		CashAmount      float64 `json:"cash_amount"`
		DeclarationDate string  `json:"declaration_date"`
		DividendType    string  `json:"dividend_type"`
		ExDividendDate  string  `json:"ex_dividend_date"`
		Frequency       int     `json:"frequency"`
		PayDate         string  `json:"pay_date"`
		RecordDate      string  `json:"record_date"`
		Ticker          string  `json:"ticker"`
	} `json:"results"`
	Status string `json:"status"`
}

type PolygonPositionData struct {
	NextPayment int     `json:"nextpayment" bson:"nextpayment"`
	PrevPayment int     `json:"prevpayment" bson:"prevpayment"`
	ExDividend  int     `json:"exdividend" bson:"exdividend"`
	CashAmount  float64 `json:"cashamount" bson:"cashamount"`
	DivPaid     int     `json:"divpaid" bson:"divpaid"`
}

type StockUtils struct {
	Ticker    string                         `json:"ticker" bson:"ticker"`
	StockList map[string]PolygonPositionData `json:"stockList" bson:"stockList"`
}

type ForexResponse struct {
	Adjusted   bool   `json:"adjusted"`
	QueryCount int    `json:"queryCount"`
	RequestID  string `json:"request_id"`
	Results    []struct {
		T  string  `json:"T"`
		C  float64 `json:"c"`
		H  float64 `json:"h"`
		L  float64 `json:"l"`
		N  int     `json:"n"`
		O  float64 `json:"o"`
		Tt int64   `json:"t"`
		V  int     `json:"v"`
		Vw float64 `json:"vw"`
	} `json:"results"`
	ResultsCount int    `json:"resultsCount"`
	Status       string `json:"status"`
	Ticker       string `json:"ticker"`
}

func CheckTickerAvailabilty(ticker string) bool {
	var tickerCheck bool
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API")
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/dividends?ticker=%s&limit=2&apiKey=%s", ticker, apiKey)

	var currStocksInDb StockUtils

	stockDb := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	err := stockDb.FindOne(context.TODO(), filter).Decode(&currStocksInDb)
	if err != nil {
		log.Println("func CheckTickerAvailabilty: ", err)
	}

	for tickerIter := range currStocksInDb.StockList {
		if tickerIter == ticker {
			tickerCheck = true
			return tickerCheck
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println("func CheckTickerAvailabilty: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("func CheckTickerAvailabilty: ", err)
	}
	var response PolygonJson
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("func CheckTickerAvailabilty: ", err)
	}

	if len(response.Results) < 1 {
		tickerCheck = false
	} else {
		tickerCheck = true
		AddTickerToDb(ticker)
	}

	return tickerCheck
}

func AddTickerToDb(ticker string) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API")
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/dividends?ticker=%s&limit=2&apiKey=%s", ticker, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}
	var response PolygonJson
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}

	t, err := time.Parse("2006-01-02", response.Results[0].PayDate)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}
	convertedNextPayment := int(t.Unix())
	t, err = time.Parse("2006-01-02", response.Results[1].PayDate)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}
	convertedPrevPayment := int(t.Unix())
	t, err = time.Parse("2006-01-02", response.Results[0].ExDividendDate)
	if err != nil {
		log.Println("func AddTickerToDb: ", err)
	}
	convertedExDivDate := int(t.Unix())

	var divPaidBool int
	if convertedNextPayment < int(time.Now().Unix()) {
		// paid out
		divPaidBool = 1
	} else {
		// yet to be paid
		divPaidBool = 0
	}

	toSendStock := PolygonPositionData{
		NextPayment: convertedNextPayment,
		PrevPayment: convertedPrevPayment,
		ExDividend:  convertedExDivDate,
		CashAmount:  response.Results[0].CashAmount,
		DivPaid:     divPaidBool,
	}

	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	update := bson.M{"$set": bson.M{"stockList." + ticker: toSendStock}}

	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("func AddTickerToDb; Failed to update document: ", err)
		return
	}
}

func GetTimestamps(ticker string, username string) {
	stockUtilsCollection := MongoClient.Database("users").Collection("stockUtils")
	userCollection := MongoClient.Database("users").Collection(username)

	stockListFilter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	userFilter := bson.M{"ticker": "positions"}
	var stockListDb StockUtils
	var userPositions Positions

	err := stockUtilsCollection.FindOne(context.TODO(), stockListFilter).Decode(&stockListDb)
	if err != nil {
		log.Println("func GetTimestamps: ", err)
	}
	err = userCollection.FindOne(context.TODO(), userFilter).Decode(&userPositions)
	if err != nil {
		log.Println("func GetTimestamps: ", err)
	}

	updateFilter := bson.M{"stocks.ticker": ticker}
	updateDates := bson.M{
		"$set": bson.M{
			"stocks.$.currency":         "USD",
			"stocks.$.divquarterlyrate": stockListDb.StockList[ticker].CashAmount,
			"stocks.$.nextpayment":      stockListDb.StockList[ticker].NextPayment,
			"stocks.$.prevpayment":      stockListDb.StockList[ticker].PrevPayment,
			"stocks.$.divpaid":          stockListDb.StockList[ticker].DivPaid,
			"stocks.$.exdivdate":        stockListDb.StockList[ticker].ExDividend,
		},
	}

	_, err = userCollection.UpdateOne(context.Background(), updateFilter, updateDates)
	if err != nil {
		log.Println("func GetTimestamps: ", err)
	}

}

func PolygonTickerUpdate(ticker string) (nextpayment int, exdivdate int, cashamount float64) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API")
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/dividends?ticker=%s&limit=2&apiKey=%s", ticker, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("func PolygonTickerUpdate: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("func PolygonTickerUpdate: ", err)
	}
	var response PolygonJson
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("func PolygonTickerUpdate: ", err)
	}

	t, err := time.Parse("2006-01-02", response.Results[0].PayDate)
	if err != nil {
		log.Println("func PolygonTickerUpdate: ", err)
	}
	convertedNextPayment := int(t.Unix())
	t, err = time.Parse("2006-01-02", response.Results[0].ExDividendDate)
	if err != nil {
		log.Println("func PolygonTickerUpdate: ", err)
	}
	convertedExDivDate := int(t.Unix())

	return convertedNextPayment, convertedExDivDate, response.Results[0].CashAmount

}
