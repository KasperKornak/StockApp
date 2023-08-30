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

// used to decode dividend api data
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

// used to store companies data in stockUtils
type PolygonPositionData struct {
	NextPayment int     `json:"nextpayment" bson:"nextpayment"`
	PrevPayment int     `json:"prevpayment" bson:"prevpayment"`
	ExDividend  int     `json:"exdividend" bson:"exdividend"`
	CashAmount  float64 `json:"cashamount" bson:"cashamount"`
	DivPaid     int     `json:"divpaid" bson:"divpaid"`
}

// used to aggregate all PolygonpositionData
type StockUtils struct {
	Ticker    string                         `json:"ticker" bson:"ticker"`
	StockList map[string]PolygonPositionData `json:"stockList" bson:"stockList"`
}

// used to decode forex response from Polygon
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
	// init variables
	var tickerCheck bool
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API")
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/dividends?ticker=%s&limit=2&apiKey=%s", ticker, apiKey)

	// first, check if ticker is already exists in stockUtils
	var currStocksInDb StockUtils
	stockDb := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	err := stockDb.FindOne(context.TODO(), filter).Decode(&currStocksInDb)
	if err != nil {
		log.Println("func CheckTickerAvailabilty: ", err)
	}

	// iterate over all tickers - if ticker is already in mongodb, return true
	for tickerIter := range currStocksInDb.StockList {
		if tickerIter == ticker {
			tickerCheck = true
			return tickerCheck
		}
	}

	// ticker wasn't in mongodb - send request to polygon
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

	// if response is empty - return false, else add new ticker to database
	if len(response.Results) < 1 {
		tickerCheck = false
	} else {
		tickerCheck = true
		AddTickerToDb(ticker)
	}

	return tickerCheck
}

// add new company to stockUtils collection
func AddTickerToDb(ticker string) {
	// read .env file, send request to polygon for data
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

	// get next, previous payment, ex-div dates; format them to unix datetime format
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

	// see if dividend has been paid out recently
	var divPaidBool int
	if convertedNextPayment < int(time.Now().Unix()) {
		// paid out
		divPaidBool = 1
	} else {
		// yet to be paid
		divPaidBool = 0
	}

	// send this struct to mongodb
	toSendStock := PolygonPositionData{
		NextPayment: convertedNextPayment,
		PrevPayment: convertedPrevPayment,
		ExDividend:  convertedExDivDate,
		CashAmount:  response.Results[0].CashAmount,
		DivPaid:     divPaidBool,
	}

	// add document to stockUtils collection
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	update := bson.M{"$set": bson.M{"stockList." + ticker: toSendStock}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println("func AddTickerToDb; Failed to update document: ", err)
		return
	}
}

// fill timestamps in users' positions document
func GetTimestamps(ticker string, username string) {
	// open connections to user and stockUtils collections, init variables
	stockUtilsCollection := MongoClient.Database("users").Collection("stockUtils")
	userCollection := MongoClient.Database("users").Collection(username)
	stockListFilter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	userFilter := bson.M{"ticker": "positions"}
	var stockListDb StockUtils
	var userPositions Positions

	// retrieve users and stockUtils documents with companies/posiitons
	err := stockUtilsCollection.FindOne(context.TODO(), stockListFilter).Decode(&stockListDb)
	if err != nil {
		log.Println("func GetTimestamps: ", err)
	}
	err = userCollection.FindOne(context.TODO(), userFilter).Decode(&userPositions)
	if err != nil {
		log.Println("func GetTimestamps: ", err)
	}

	// update the fields in users position list with retrieved timestamps
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

// used to get latest dividend-payment related data
func PolygonTickerUpdate(ticker string) (nextpayment int, exdivdate int, cashamount float64) {
	// send request to polygon api for dividend data
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

	// convert received data to unix datetime format
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
	log.Println("ticker: ", ticker, "nextPayment: ", convertedNextPayment, "; exDivDate: ", convertedExDivDate)
	// return next payment and ex-dividend date and cash amount
	return convertedNextPayment, convertedExDivDate, response.Results[0].CashAmount
}
