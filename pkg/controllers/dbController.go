package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/KasperKornak/StockApp/pkg/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetStocks(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	stockSlice := models.GetStocks(Client)

	res, _ := json.Marshal(stockSlice)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetStockByTicker(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	vars := mux.Vars(r)
	ticker := vars["ticker"]
	stock := models.GetStockByTicker(ticker, Client)
	res, _ := json.Marshal(stock)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

}
