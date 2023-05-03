package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Company struct {
	Ticker           string  `json:"ticker"`
	Shares           int     `json:"shares"`
	Domestictax      int     `json:"domestictax"`
	Currency         string  `json:"currency"`
	DivQuarterlyRate float64 `json:"divquarterlyrate" bson:"divquarterlyrate"`
	DivYTD           float64 `json:"divytd" bson:"divytd"`
	DivPLN           float64 `json:"divpln" bson:"divpln"`
	NextPayment      int     `json:"nextpayment" bson:"nextpayment"`
	PrevPayment      int     `json:"prevpayment" bson:"prevpayment"`
}

type DeleteTicker struct {
	DeleteSymbol string `json:"symbol"`
}

func ModelGetStocks(Client *mongo.Client) []Company {
	stocks := Client.Database("stock").Collection("tickers")

	filter := bson.M{"_id": bson.M{"$exists": true}, "shares": bson.M{"$exists": true}, "domestictax": bson.M{"$exists": true}, "currency": bson.M{"$exists": true}, "divquarterlyrate": bson.M{"$exists": true}}
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

func ModelGetStockByTicker(ticker string, Client *mongo.Client) Company {
	stocks := Client.Database("stock").Collection("tickers")
	var Stock Company

	filter := bson.M{"ticker": ticker}
	err := stocks.FindOne(context.TODO(), filter).Decode(&Stock)

	if err != nil {
		log.Fatal(err)
	}

	return Stock
}

func ModelDeletePosition(ticker string, Client *mongo.Client) error {
	stocks := Client.Database("stock").Collection("tickers")
	filter := bson.M{"ticker": ticker}
	result, err := stocks.DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("error deleting document: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func ModelCreatePosition(ticker string, shares int, domestictax int, currency string, divQuarterlyRate float64, divytd float64, divpln float64, nextpayment int, prevpayment int, Client *mongo.Client) error {
	stocks := Client.Database("stock").Collection("tickers")
	newPosition := &Company{
		Ticker:           ticker,
		Shares:           shares,
		Domestictax:      domestictax,
		Currency:         currency,
		DivQuarterlyRate: divQuarterlyRate,
		DivYTD:           divytd,
		DivPLN:           divpln,
		NextPayment:      nextpayment,
		PrevPayment:      prevpayment,
	}

	_, err := stocks.InsertOne(context.TODO(), newPosition)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func ModelUpdatePosition(ticker string, shares int, domestictax int, currency string, divQuarterlyRate float64, divytd float64, divpln float64, nextpayment int, prevpayment int, Client *mongo.Client) error {
	stocks := Client.Database("stock").Collection("tickers")

	currentStatus := ModelGetStockByTicker(ticker, Client)

	var updateStock Company
	updateStock.Ticker = ticker

	if shares != currentStatus.Shares {
		updateStock.Shares = shares
	} else {
		updateStock.Shares = currentStatus.Shares
	}

	if domestictax != currentStatus.Domestictax {
		updateStock.Domestictax = domestictax
	} else {
		updateStock.Domestictax = currentStatus.Domestictax
	}

	if currency != currentStatus.Currency {
		updateStock.Currency = currency
	} else {
		updateStock.Currency = currentStatus.Currency
	}

	if divQuarterlyRate != currentStatus.DivQuarterlyRate {
		updateStock.DivQuarterlyRate = divQuarterlyRate
	} else {
		updateStock.DivQuarterlyRate = currentStatus.DivQuarterlyRate
	}

	if divytd != currentStatus.DivYTD {
		updateStock.DivYTD = divytd
	} else {
		updateStock.DivYTD = currentStatus.DivYTD
	}

	if divpln != currentStatus.DivPLN {
		updateStock.DivPLN = divpln
	} else {
		updateStock.DivPLN = currentStatus.DivPLN
	}

	if nextpayment != currentStatus.NextPayment {
		updateStock.NextPayment = nextpayment
	} else {
		updateStock.NextPayment = currentStatus.NextPayment
	}

	if prevpayment != currentStatus.PrevPayment {
		updateStock.PrevPayment = prevpayment
	} else {
		updateStock.PrevPayment = currentStatus.PrevPayment
	}

	filter := bson.M{"ticker": ticker}
	update := bson.M{"$set": updateStock}
	_, err := stocks.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
