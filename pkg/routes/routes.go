package routes

import (
	"github.com/KasperKornak/StockApp/pkg/controllers"
	"github.com/gorilla/mux"
)

var RegisterStocks = func(router *mux.Router) {
	router.HandleFunc("/stocks", controllers.GetStocks).Methods("GET")
	router.HandleFunc("/stock/{ticker}", controllers.GetStockByTicker).Methods("GET")
	router.HandleFunc("/delete", controllers.DeletePosition).Methods("DELETE")
	router.HandleFunc("/create", controllers.CreatePosition).Methods("POST")
	router.HandleFunc("/update", controllers.UpdatePosition).Methods("PUT")
	router.HandleFunc("/home", controllers.StocksHTML).Methods("GET")
}
