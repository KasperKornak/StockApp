package models

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *redis.Client
var MongoClient *mongo.Client

func Init() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	MongoClient, _ = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
}
