package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/KasperKornak/StockApp/pkg/controllers"
	"github.com/KasperKornak/StockApp/pkg/middleware"
	"github.com/KasperKornak/StockApp/pkg/models"
	"github.com/KasperKornak/StockApp/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	Client := config.MongoConnect()
	defer Client.Disconnect(context.TODO())

	// check if deleted stocks document exists
	// if not - create one
	tickers := Client.Database("stock").Collection("tickers")
	filter := bson.M{"year": time.Now().Year(), "ticker": "DELETED_SUM"}
	err := tickers.FindOne(context.TODO(), filter).Err()
	if err == mongo.ErrNoDocuments {
		var newDocument models.DeletedCompany
		newDocument.Year = time.Now().Year()
		newDocument.Ticker = "DELETED_SUM"
		newDocument.DivYTD = 0.0
		newDocument.DivPLN = 0.0

		_, err := tickers.InsertOne(context.TODO(), &newDocument)
		if err != nil {
			panic(err)
		}
	}

	// check if year summary document exists
	// if not - create one
	filter = bson.M{"year": time.Now().Year(), "ticker": "YEAR_SUMMARY"}
	err = tickers.FindOne(context.TODO(), filter).Err()
	if err == mongo.ErrNoDocuments {
		var newDocument models.DeletedCompany
		newDocument.Year = time.Now().Year()
		newDocument.Ticker = "YEAR_SUMMARY"
		newDocument.DivYTD = 0.0
		newDocument.DivPLN = 0.0

		_, err := tickers.InsertOne(context.TODO(), &newDocument)
		if err != nil {
			panic(err)
		}
	}

	controllers.UpdateSummary()

}

func main() {
	// open connection to mongodb
	Client := config.MongoConnect()
	defer Client.Disconnect(context.TODO())

	// run the app
	r := mux.NewRouter()
	r.Use(middleware.MongoMiddleware(Client))
	routes.RegisterStocks(r)

	http.Handle("/", r)

	// check for new dividend payment dates and
	c := cron.New()
	c.AddFunc("53 22 * * *", func() {
		fmt.Println("Starting sync..")
		controllers.GetPaymentDate()
		controllers.CheckPayment()
		controllers.UpdateSummary()
		fmt.Println("Sync done 🦖")
	})

	c.Start()

	log.Fatal(http.ListenAndServe(":9010", r))
}
