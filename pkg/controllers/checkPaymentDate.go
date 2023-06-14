package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Response struct {
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

func GetPaymentDate() {
	Client := config.MongoConnect()
	stockSlice := ModelGetStocks(Client)
	tickers := GetMongoTickers()
	apiNum := 1

	for _, stock := range tickers {
		newestData := SendAPIGet(stock)
		for _, company := range stockSlice {
			if (stock == company.Ticker) && (newestData > company.NextPayment) {
				next := company.NextPayment
				filter := bson.M{"ticker": company.Ticker}
				stocks := Client.Database("stock").Collection("tickers")
				update := bson.M{
					"$set": bson.M{
						"nextpayment": newestData,
						"prevpayment": next,
					},
				}
				_, err := stocks.UpdateOne(context.TODO(), filter, update)
				if err != nil {
					panic(err)
				}
				fmt.Printf("Updated dividend for: %s\n", company.Ticker)
			}
			apiNum += 1
			if apiNum%5 == 0 {
				time.Sleep(1 * time.Minute)
			}
		}
	}
	err := Client.Disconnect(context.TODO())
	if err != nil {
		panic(err)
	}
}

func CheckPayment() {
	time.Sleep(1 * time.Minute)
	Client := config.MongoConnect()
	stockSlice := ModelGetStocks(Client)
	apiNum := 1
	for _, company := range stockSlice {
		if (company.NextPayment <= int(time.Now().Unix())) && (company.NextPayment != company.PrevPayment) {
			pair := fmt.Sprintf("C:%sPLN", company.Currency)
			q := GetForex(pair)
			noShares := company.Shares
			div := company.DivQuarterlyRate
			divPLNtoSend := div * float64(noShares) * q * float64(company.Domestictax) / 100.0

			var divUSDtoSend float64
			if company.Currency != "USD" {
				apiNum += 1
				pairCorr := fmt.Sprintf("C:%sUSD", company.Currency)
				correction := GetForex(pairCorr)

				divUSDtoSend = div * float64(noShares) * correction
			} else {
				divUSDtoSend = div * float64(noShares)
			}
			filter := bson.M{"ticker": company.Ticker}
			stocks := Client.Database("stock").Collection("tickers")
			update := bson.M{"$set": bson.M{"divytd": (company.DivYTD + divUSDtoSend)}}
			_, err := stocks.UpdateOne(context.TODO(), filter, update)
			if err != nil {
				panic(err)
			}

			updated := bson.M{"$set": bson.M{"divpln": (company.DivPLN + divPLNtoSend)}}
			_, err = stocks.UpdateOne(context.TODO(), filter, updated)
			if err != nil {
				panic(err)
			}

			updateNextDate := bson.M{"$set": bson.M{"prevpayment": (company.NextPayment)}}
			_, err = stocks.UpdateOne(context.TODO(), filter, updateNextDate)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Recieved dividend from: %s\n", company.Ticker)
			fmt.Printf("Amount in USD: %f\n", divUSDtoSend)
			fmt.Printf("Tax to pay in PLN: %f\n", divPLNtoSend)
			SlackRequest(company.Ticker, divUSDtoSend, divPLNtoSend)
		}
		apiNum += 1
		if apiNum%5 == 0 {
			time.Sleep(1 * time.Minute)
		}
	}
	err := Client.Disconnect(context.TODO())
	if err != nil {
		panic(err)
	}
}

func UpdateSummary() {
	var updatedDoc DeletedCompany
	updatedDoc.Year = time.Now().Year()
	updatedDoc.Ticker = "YEAR_SUMMARY"
	updatedDoc.DivYTD = 0.0
	updatedDoc.DivPLN = 0.0

	Client := config.MongoConnect()
	tickers := Client.Database("stock").Collection("tickers")
	stockSlice := ModelGetStocks(Client)
	divTax := 0.0
	divRec := 0.0

	for _, stock := range stockSlice {
		divTax = divTax + stock.DivPLN
		divRec = divRec + stock.DivYTD
	}

	deletedStocks := ModelGetStockByTicker("DELETED_SUM", Client)

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
		var deleted DeletedCompany
		deleted.Year = time.Now().Year()
		deleted.Ticker = "DELETED_SUM"
		deleted.DivYTD = 0.0
		deleted.DivPLN = 0.0

		_, err := tickers.InsertOne(context.TODO(), &deleted)
		if err != nil {
			panic(err)
		}

		var newDocument DeletedCompany
		newDocument.Year = time.Now().Year()
		newDocument.Ticker = "YEAR_SUMMARY"
		newDocument.DivYTD = 0.0
		newDocument.DivPLN = 0.0

		_, err = tickers.InsertOne(context.TODO(), &newDocument)
		if err != nil {
			panic(err)
		}

		stockSlice := ModelGetStocks(Client)
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

func GetMongoTickers() []string {
	var tickerSlice []string
	Client := config.MongoConnect()
	tickers := Client.Database("stock").Collection("tickers")
	defer Client.Disconnect(context.TODO())

	filter := bson.M{"_id": bson.M{"$exists": true}, "shares": bson.M{"$exists": true}, "domestictax": bson.M{"$exists": true}, "currency": bson.M{"$exists": true}, "divquarterlyrate": bson.M{"$exists": true}}
	cur, err := tickers.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
	}

	var result struct {
		Ticker string `bson:"ticker"`
	}

	for cur.Next(context.TODO()) {
		cur.Decode(&result)

		if result.Ticker != "YEAR_SUMMARY" && result.Ticker != "DELETED_SUM" {
			tickerSlice = append(tickerSlice, result.Ticker)
		}
	}

	return tickerSlice
}

func SendAPIGet(tickerToCheck string) int {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API_KEY")
	url := fmt.Sprintf("https://api.polygon.io/v3/reference/dividends?ticker=%s&limit=1&apiKey=%s", tickerToCheck, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	tempDivDate := response.Results[0].PayDate
	t, _ := time.Parse("2006-01-02", tempDivDate)

	divDate := int(t.Unix())

	return divDate
}

func GetForex(pair string) float64 {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API_KEY")
	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", pair, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var response ForexResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
	}

	exRate := response.Results[0].C

	return exRate
}
