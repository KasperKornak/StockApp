package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/KasperKornak/StockApp/pkg/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetStocks(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	stockSlice := models.ModelGetStocks(Client)

	res, _ := json.Marshal(stockSlice)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetStockByTicker(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	vars := mux.Vars(r)
	ticker := vars["ticker"]
	stock := models.ModelGetStockByTicker(ticker, Client)
	res, _ := json.Marshal(stock)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

}

func DeletePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body models.DeleteTicker
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker := body.DeleteSymbol
	_ = models.ModelDeletePosition(ticker, Client)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(body.DeleteSymbol))
}

func CreatePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body models.Company
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker, shares, domestictax := body.Ticker, body.Shares, body.Domestictax
	_ = models.ModelCreatePosition(ticker, shares, domestictax, Client)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

func UpdatePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body models.Company
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker, shares, domestictax := body.Ticker, body.Shares, body.Domestictax
	_ = models.ModelUpdatePosition(ticker, shares, domestictax, Client)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)

}
