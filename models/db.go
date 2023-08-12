package models

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// global variable for mongodb and local for redis
var client *redis.Client
var MongoClient *mongo.Client

// init func which initializes connections to databases
func Init() {
	// open connection to redis
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// read .env file and open connection to mongdodb
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	MongoClient, _ = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))

	// func to close connection
	registerDisconnectHandler()
}

// handler used to close connection to mongo when app process is killed
func registerDisconnectHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		MongoClient.Disconnect(context.Background())
		os.Exit(0)
	}()
}
