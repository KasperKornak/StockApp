package main

import (
	"context"
	"log"
	"net/http"

	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/KasperKornak/StockApp/pkg/controllers"
	"github.com/KasperKornak/StockApp/pkg/middleware"
	"github.com/KasperKornak/StockApp/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
)

func main() {
	controllers.GetPaymentDate()
	controllers.CheckPayment()
	Client := config.MongoConnect()
	defer Client.Disconnect(context.TODO())

	r := mux.NewRouter()
	r.Use(middleware.MongoMiddleware(Client))
	routes.RegisterStocks(r)

	http.Handle("/", r)

	// create a new cron job
	c := cron.New()

	// add the job to the cron scheduler
	c.AddFunc("0 13 * * *", func() {
		// execute the required functions
		controllers.GetPaymentDate()
		controllers.CheckPayment()
	})

	// start the cron scheduler
	c.Start()

	log.Fatal(http.ListenAndServe(":9010", r))
}
