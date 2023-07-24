package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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
		log.Println(err)
	}

	return userList
}

func UpdateUserList() {
	currUserList := RetrieveUsers()
	currAvailableStocks := RetrieveAvailableStocks()
	for _, username := range currUserList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions

		curUserFilter := bson.M{"ticker": "positions"}
		_ = currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)

		for _, position := range currUserStocks.Stocks {
			for availableStockTicker, availableStock := range currAvailableStocks.StockList {
				if position.Ticker == availableStockTicker {
					if position.NextPayment != availableStock.NextPayment {
						position.NextPayment = availableStock.NextPayment
						position.PrevPayment = availableStock.PrevPayment
						position.DivPaid = availableStock.DivPaid
						position.ExDivDate = availableStock.ExDividend
					}

				}
			}
		}
		update := bson.M{
			"$set": bson.M{
				"stocks": currUserStocks.Stocks,
			},
		}
		_, err := currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println(err)
		}
	}

}

func RetrieveAvailableStocks() StockUtils {
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}

	var stockList StockUtils
	err := collection.FindOne(context.TODO(), filter).Decode(&stockList)

	if err != nil {
		log.Println(err)
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
		if stockData.NextPayment < int(time.Now().Unix()) {
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
			if err != nil {
				log.Println(err)
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
	i := 1
	for _, username := range userList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions
		currentMonth := time.Now().Month()
		curUserFilter := bson.M{"ticker": "positions"}
		_ = currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		var months initMongoMonths
		_ = currUserCollection.FindOne(context.TODO(), bson.M{"ticker": "MONTH_SUMARY"}).Decode(&months)
		for _, position := range currUserStocks.Stocks {
			if (position.DivPaid == 0) && (position.NextPayment <= int(time.Now().Unix())) {
				position.DivYTD = float64(position.SharesAtExDiv)*position.DivQuarterlyRate + position.DivYTD
				position.DivPLN = position.DivQuarterlyRate * float64(position.SharesAtExDiv) * currencyPair * float64(position.Domestictax) / 100.0
				for i, month := range months.Months {
					if month.Name[:3] == currentMonth.String()[:3] {
						months.Months[i].Value = position.DivYTD
						break
					}
				}
			}

			if i%5 == 0 {
				time.Sleep(60 * time.Second)
				i = 0
			}
		}
		update := bson.M{
			"$set": bson.M{
				"stocks": currUserStocks.Stocks,
			},
		}
		_, err := currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println(err)
		}
		update = bson.M{"$set": months}
		_, err = currUserCollection.UpdateOne(context.TODO(), bson.M{"ticker": "MONTH_SUMARY"}, update)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetForex() float64 {
	// if err := godotenv.Load(); err != nil {
	// 	fmt.Println("No .env file found")
	// }
	// apiKey := os.Getenv("POLYGON_API_KEY")
	url := "https://api.polygon.io/v2/aggs/ticker/C:USDPLN/prev?adjusted=true&apiKey=GBU7kTsNoJM2iUN7iMzOZFnyPxz3Ty_a"

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
	log.Println(response)
	exRate := response.Results[0].C

	return exRate
}
