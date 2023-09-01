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

// username list retrieval
type UsernamesDocument struct {
	ID        string   `json:"_id" bson:"_id"`
	Usernames []string `json:"usernames" bson:"usernames"`
}

// retrieve usernames from mongodb
func RetrieveUsers() UsernamesDocument {
	// init variables and connection to mongodb
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "ALL_USERNAMES"}
	var userList UsernamesDocument

	// retrieve username list
	err := collection.FindOne(context.TODO(), filter).Decode(&userList)
	if err != nil {
		log.Println("func RetrieveUsers: ", err)
	}

	return userList
}

// used to update positions in users' collection
func UpdateUserList() {
	// init variables, retrieve username list and all tracked stocks
	currUserList := RetrieveUsers()
	currAvailableStocks := RetrieveAvailableStocks()

	// iterate over username list, connect to their collection
	for _, username := range currUserList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions
		curUserFilter := bson.M{"ticker": "positions"}
		err := currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		if err != nil {
			log.Println("func UpdateUserList:", err)
		}

		// iterate over positions of user and available stocks
		// if tickers match, start comparison
		for _, position := range currUserStocks.Stocks {
			for availableStockTicker, availableStock := range currAvailableStocks.StockList {
				// update user's dividend-related data on position list
				if (position.Ticker == availableStockTicker) && (position.Ticker != "DELETED_SUM") {
					if position.NextPayment != availableStock.NextPayment {
						position.NextPayment = availableStock.NextPayment
						position.PrevPayment = availableStock.PrevPayment
						position.ExDivDate = availableStock.ExDividend
						log.Println("Updated position for: ", username, " stock updated: ", position.Ticker)
					} else {
						log.Println("Didn't update position for: ", username, " stock updated: ", position.Ticker)
					}
				}
			}
		}
		// update user's positions
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

// used to retrieve all tracked tickers and their data
func RetrieveAvailableStocks() StockUtils {
	// init variables and connection to mongodb
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	var stockList StockUtils

	// retrieve all tracked tickers and their data
	err := collection.FindOne(context.TODO(), filter).Decode(&stockList)
	if err != nil {
		log.Println("func RetrieveAvailableStocks: ", err)
	}

	return stockList
}

// run this func last in cronjob
// updates dividend-related data in tracked tickers document
func UpdateStockDb() {
	// init variables, add counter to bypass polygon api restrictions, connect to mongodb
	availableStocks := RetrieveAvailableStocks()
	timeCounter := 0
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}

	// iterate over available stocks
	for ticker, stockData := range availableStocks.StockList {
		// get next payment, ex-dividend dates and dividend amount
		nextpayment, exdivdate, cashamount := PolygonTickerUpdate(ticker)
		var divPaidBool int
		if nextpayment < int(time.Now().Unix()) {
			// paid out
			divPaidBool = 1
		} else {
			// yet to be paid
			divPaidBool = 0
		}

		// set values of updated ticker fields
		if stockData.NextPayment < nextpayment {
			updatedPosition := PolygonPositionData{
				NextPayment: nextpayment,
				PrevPayment: stockData.NextPayment,
				ExDividend:  exdivdate,
				CashAmount:  cashamount,
				DivPaid:     divPaidBool,
			}

			// send update together with small debugging message
			update := bson.M{"$set": bson.M{fmt.Sprintf("stockList.%s", ticker): updatedPosition}}
			_, err := collection.UpdateOne(context.TODO(), filter, update)
			log.Println("Updated data for: ", stockData, ";\n Next payment: ", stockData.NextPayment, ";\n Prevoius Payment: ",
				stockData.PrevPayment, ";\n Ex-div date: ", stockData.ExDividend, ";\n Cash amount: ", stockData.CashAmount,
				";\n Div paid: ", stockData.DivPaid)
			if err != nil {
				log.Println("func UpdateStockDb: ", err)
			}
		}
		// counter control - if counter == 5, sleep for one minute
		if (timeCounter % 5) == 0 {
			time.Sleep(60 * time.Second)
			timeCounter = 0
		}
	}
}

// runs on each user individually, checks if dividend has been paid out and updates paid out data
func CalculateDividends() {
	// init variables, get current exchange rate between USD and PLN
	userList := RetrieveUsers()
	// TODO: use official NBP api insead of Polygon data
	currencyPair := GetForex()

	// iterate over usernames, init variables
	for _, username := range userList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions

		// get position documents, month summary
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

		// add dividend paid, tax to positions and month summary
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

		// update the stocks array in the document
		update := bson.M{"$set": bson.M{"stocks": currUserStocks.Stocks}}
		_, err = currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}

		// update the MONTH_SUMARY document with modified months
		update = bson.M{"$set": months}
		_, err = currUserCollection.UpdateOne(context.TODO(), bson.M{"ticker": "MONTH_SUMARY"}, update)
		if err != nil {
			log.Println("func CalculateDividends: ", err)
		}
		UpdateSummary(username)
	}
}

// get usd/pln pair data from polygon
func GetForex() float64 {
	// load polygon api key, send request
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

	// get exchange rate
	exRate := response.Results[0].C

	return exRate
}

// returns absolute value of a number
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// updates the summary documents
func UpdateSummary(username string) {
	// init variables and connection to mongodb
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

	// variables used to add all values of DivPLN and DivYTD
	divYtd := 0.0
	divPln := 0.0

	for _, position := range positions.Stocks {
		divPln += position.DivPLN
		divYtd += position.DivYTD
	}

	// update the summary struct with divPln and divYtd values
	summary.DividendTax = divPln
	summary.DividendsYTD = divYtd
	update := bson.M{"$set": summary}
	_, err = collection.UpdateOne(context.TODO(), yearFilter, update)
	if err != nil {
		log.Println("func UpdateSummary: ", err)
	}
}
