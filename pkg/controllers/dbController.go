package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetStocks(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	stockSlice := ModelGetStocks(Client)

	res, _ := json.Marshal(stockSlice)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetStockByTicker(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	vars := mux.Vars(r)
	ticker := vars["ticker"]
	stock := ModelGetStockByTicker(ticker, Client)
	res, _ := json.Marshal(stock)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func DeletePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body DeleteTicker
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker := body.DeleteSymbol

	tempBody := ModelGetStockByTicker(ticker, Client)

	divRec, divTax := tempBody.DivYTD, tempBody.DivPLN
	_ = TransferDivs(divRec, divTax, Client)

	_ = ModelDeletePosition(ticker, Client)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(body.DeleteSymbol))
}

func CreatePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body Company
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker, shares, domestictax := body.Ticker, body.Shares, body.Domestictax
	currency, divQuarterlyRate, prevpayment := body.Currency, body.DivQuarterlyRate, body.PrevPayment
	divytd, divpln, nextpayment := body.DivYTD, body.DivPLN, body.NextPayment
	_ = ModelCreatePosition(ticker, shares, domestictax, currency, divQuarterlyRate, divytd, divpln, nextpayment, prevpayment, Client)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
	UpdateSummary()
}

func UpdatePosition(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	var body Company
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Fatal(err)
		return
	}
	ticker, shares, domestictax := body.Ticker, body.Shares, body.Domestictax
	currency, divQuarterlyRate, prevpayment := body.Currency, body.DivQuarterlyRate, body.PrevPayment
	divytd, divpln, nextpayment := body.DivYTD, body.DivPLN, body.NextPayment
	_ = ModelUpdatePosition(ticker, shares, domestictax, currency, divQuarterlyRate, divytd, divpln, nextpayment, prevpayment, Client)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(body)
}

func StocksHTML(w http.ResponseWriter, r *http.Request) {
	Client := r.Context().Value("mongo").(*mongo.Client)
	stockSlice := ModelGetStocks(Client)

	res, err := json.Marshal(stockSlice)
	if err != nil {
		http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
		return
	}

	var data []Company
	err = json.Unmarshal(res, &data)
	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusInternalServerError)
		return
	}

	var deletedCompanies DeletedCompany
	collection := Client.Database("stock").Collection("tickers")
	err = collection.FindOne(context.TODO(), bson.M{"ticker": "DELETED_SUM", "year": time.Now().Year()}).Decode(&deletedCompanies)
	if err != nil {
		log.Fatal(err)
	}

	newCompany := Company{
		Ticker:           deletedCompanies.Ticker,
		Shares:           0,
		Domestictax:      0,
		Currency:         "",
		DivQuarterlyRate: 0,
		DivYTD:           deletedCompanies.DivYTD,
		DivPLN:           deletedCompanies.DivPLN,
		NextPayment:      0,
		PrevPayment:      0,
	}

	tmpl, err := template.ParseFiles("../pkg/template/table.tmpl")
	if err != nil {
		http.Error(w, "Error rendering HTML template", http.StatusInternalServerError)
		return
	}
	data = append(data, newCompany)
	tmpl.Execute(w, data)

}
