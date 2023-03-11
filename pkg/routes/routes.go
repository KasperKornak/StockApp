package routes

import (
	"github.com/KasperKornak/StockApp/pkg/controllers"
	"github.com/gorilla/mux"
)

var RegisterStocks = func(router *mux.Router) {
	router.HandleFunc("/stocks/", controllers.GetStocks).Methods("GET")
	router.HandleFunc("/stock/{ticker}/", controllers.GetStockByTicker).Methods("GET")
}
