package main

import (
	"context"
	"log"
	"net/http"

	"github.com/KasperKornak/StockApp/pkg/config"
	"github.com/KasperKornak/StockApp/pkg/middleware"
	"github.com/KasperKornak/StockApp/pkg/routes"
	"github.com/gorilla/mux"
)

func main() {
	Client := config.MongoConnect()
	defer Client.Disconnect(context.TODO())

	r := mux.NewRouter()
	r.Use(middleware.MongoMiddleware(Client))
	routes.RegisterStocks(r)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":9010", r))
}
