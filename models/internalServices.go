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

type UsernamesDocument struct {
	ID        string   `json:"_id" bson:"_id"`
	Usernames []string `json:"usernames" bson:"usernames"`
}

func RetrieveUsers() UsernamesDocument {
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "ALL_USERNAMES"}

	var userList UsernamesDocument
	err := collection.FindOne(context.TODO(), filter).Decode(&userList)

	if err != nil {
		log.Println("func RetrieveUsers: ", err)
	}

	return userList
}

func UpdateUserList() {
	currUserList := RetrieveUsers()
	currentTime := int(time.Now().Unix())
	currAvailableStocks := RetrieveAvailableStocks()
	for _, username := range currUserList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions

		curUserFilter := bson.M{"ticker": "positions"}
		err := currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		if err != nil {
			log.Println("func UpdateUserList:", err)
		}

		for _, position := range currUserStocks.Stocks {
			for availableStockTicker, availableStock := range currAvailableStocks.StockList {
				if (position.Ticker == availableStockTicker) && (position.Ticker != "DELETED_SUM") {
					if position.NextPayment != availableStock.NextPayment {
						position.NextPayment = availableStock.NextPayment
						position.PrevPayment = availableStock.PrevPayment
						position.DivPaid = availableStock.DivPaid
						position.ExDivDate = availableStock.ExDividend
					}
					if Abs(position.ExDivDate-currentTime) <= 48*60*60 {
						position.SharesAtExDiv = position.Shares
					} else {
						position.SharesAtExDiv = position.SharesAtExDiv
					}

				}
			}
		}
		update := bson.M{
			"$set": bson.M{
				"stocks": currUserStocks.Stocks,
			},
		}
		_, err = currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println("func UpdateUserList: ", err)
		}
	}

}

func RetrieveAvailableStocks() StockUtils {
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}

	var stockList StockUtils
	err := collection.FindOne(context.TODO(), filter).Decode(&stockList)

	if err != nil {
		log.Println("func RetrieveAvailableStocks: ", err)
	}

	return stockList
}

// run this func last
func UpdateStockDb() {
	availableStocks := RetrieveAvailableStocks()
	timeCounter := 0
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	for ticker, stockData := range availableStocks.StockList {
		nextpayment, exdivdate, cashamount := PolygonTickerUpdate(ticker)

		var divPaidBool int
		if nextpayment < int(time.Now().Unix()) {
			// paid out
			divPaidBool = 1
		} else {
			// yet to be paid
			divPaidBool = 0
		}

		if stockData.NextPayment < nextpayment {
			updatedPosition := PolygonPositionData{
				NextPayment: nextpayment,
				PrevPayment: stockData.NextPayment,
				ExDividend:  exdivdate,
				CashAmount:  cashamount,
				DivPaid:     divPaidBool,
			}
			update := bson.M{"$set": bson.M{fmt.Sprintf("stockList.%s", ticker): updatedPosition}}
			_, err := collection.UpdateOne(context.TODO(), filter, update)
			log.Println("Updated data for: ", stockData, ";\n Next payment: ", stockData.NextPayment, ";\n Prevoius Payment: ",
				stockData.PrevPayment, ";\n Ex-div date: ", stockData.ExDividend, ";\n Cash amount: ", stockData.CashAmount,
				";\n Div paid: ", stockData.DivPaid)
			if err != nil {
				log.Println("func UpdateStockDb: ", err)
			}
		}
		if (timeCounter % 5) == 0 {
			time.Sleep(60 * time.Second)
			timeCounter = 0
		}
	}
}

func CalculateDividends() {
	userList := RetrieveUsers()
	currencyPair := GetForex()
	for _, username := range userList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions
		currentMonth := time.Now().Month()
		curUserFilter := bson.M{"ticker": "positions"}
		err := currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}
		var months InitMongoMonths
		err = currUserCollection.FindOne(context.TODO(), bson.M{"ticker": "MONTH_SUMARY"}).Decode(&months)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}
		for i, position := range currUserStocks.Stocks {
			if (position.Ticker != "DELETED_SUM") && (position.DivPaid == 0) && (position.NextPayment <= int(time.Now().Unix())) {

				currUserStocks.Stocks[i].DivYTD = float64(position.SharesAtExDiv)*position.DivQuarterlyRate + position.DivYTD
				currUserStocks.Stocks[i].DivPLN = position.DivQuarterlyRate*float64(position.SharesAtExDiv)*currencyPair*float64(position.Domestictax)/100.0 + position.DivPLN
				currUserStocks.Stocks[i].DivPaid = 1
				for j, month := range months.Months {
					if month.Name[:3] == currentMonth.String()[:3] {
						months.Months[j].Value = months.Months[j].Value + float64(position.SharesAtExDiv)*position.DivQuarterlyRate
						break
					}
				}
			}
		}

		// Update the stocks array in the document
		update := bson.M{"$set": bson.M{"stocks": currUserStocks.Stocks}}
		_, err = currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}

		// Update the MONTH_SUMARY document with modified months
		update = bson.M{"$set": months}
		_, err = currUserCollection.UpdateOne(context.TODO(), bson.M{"ticker": "MONTH_SUMARY"}, update)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}
		UpdateSummary(username)
	}
}

func GetForex() float64 {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	apiKey := os.Getenv("POLYGON_API")
	url := fmt.Sprintf("https://api.polygon.io/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", "C:USDPLN", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Println("func GetForex: ", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("func GetForex: ", err)
	}
	var response ForexResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("func GetForex: ", err)
	}
	log.Println(response)
	exRate := response.Results[0].C

	return exRate
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func UpdateSummary(username string) {
	collection := MongoClient.Database("users").Collection(username)
	yearFilter := bson.M{"ticker": "YEAR_SUMMARY", "year": time.Now().Year()}
	positionFilter := bson.M{"ticker": "positions"}
	var summary initYearSummary
	var positions Positions

	err := collection.FindOne(context.TODO(), yearFilter).Decode(&summary)
	if err != nil {
		log.Println("func UpdateSummary: ", err)
	}
	err = collection.FindOne(context.TODO(), positionFilter).Decode(&positions)
	if err != nil {
		log.Println("func UpdateSummary: ", err)
	}

	divYtd := 0.0
	divPln := 0.0

	for _, position := range positions.Stocks {
		divPln += position.DivPLN
		divYtd += position.DivYTD
	}

	// Update the summary struct with divPln and divYtd values
	summary.DividendTax = divPln
	summary.DividendsYTD = divYtd

	update := bson.M{"$set": summary}
	_, err = collection.UpdateOne(context.TODO(), yearFilter, update)
	if err != nil {
		log.Println("func UpdateSummary: ", err)
	}
}
