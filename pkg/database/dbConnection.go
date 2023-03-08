package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(context.TODO())

	stocks := client.Database("stock").Collection("tickers")

	cursor, err := stocks.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	// var companies []bson.M
	// if err = cursor.All(context.TODO(), &companies); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("stocks: ", companies)

	for cursor.Next(context.TODO()) {
		var cmp bson.M
		if err = cursor.Decode(&cmp); err != nil {
			log.Fatal(err)
		}
		fmt.Println("-------")
		fmt.Println(cmp)
	}
}
