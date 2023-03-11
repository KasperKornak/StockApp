package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Company struct {
	Ticker      string `json:"ticker"`
	Shares      int    `json:"shares"`
	Domestictax int    `json:"domestictax"`
}

func GetStocks(Client *mongo.Client) []Company {
	stocks := Client.Database("stock").Collection("tickers")

	filter := bson.M{"_id": bson.M{"$exists": true}, "shares": bson.M{"$exists": true}, "domestictax": bson.M{"$exists": true}}
	cur, err := stocks.Find(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}

	var StockSlice []Company
	for cur.Next(context.Background()) {
		var t Company
		err := cur.Decode(&t)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		StockSlice = append(StockSlice, t)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return nil
	}

	return StockSlice
}

func GetStockByTicker(ticker string, Client *mongo.Client) Company {
	stocks := Client.Database("stock").Collection("tickers")
	var Stock Company

	filter := bson.M{"ticker": ticker}
	err := stocks.FindOne(context.TODO(), filter).Decode(&Stock)

	if err != nil {
		log.Fatal(err)
	}

	return Stock
}
