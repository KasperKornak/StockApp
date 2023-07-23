package models

import (
	"context"
	"fmt"
	"log"
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

// func UpdateUserList(newList UsernamesDocument) {
// 	collection := MongoClient.Database("users").Collection("stockUtils")
// 	filter := bson.M{"ticker": "ALL_USERNAMES"}
// 	currList :=
// }

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

func UpdateStockDb() {
	availableStocks := RetrieveAvailableStocks()
	timeCounter := 0
	collection := MongoClient.Database("users").Collection("stockUtils")
	filter := bson.M{"ticker": "AVAILABLE_STOCKS"}
	for ticker, stockData := range availableStocks.StockList {
		nextpayment, exdivdate, cashamount := PolygonTickerUpdate(ticker)
		if stockData.NextPayment < nextpayment {
			updatedPosition := PolygonPositionData{
				NextPayment: nextpayment,
				PrevPayment: stockData.NextPayment,
				ExDividend:  exdivdate,
				CashAmount:  cashamount,
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
