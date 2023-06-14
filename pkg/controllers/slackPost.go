package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type SlackPost struct {
	Ticker string  `json:"ticker"`
	DivUSD float64 `json:"divUSD"`
	DivPLN float64 `json:"divPLN"`
}

func SlackRequest(ticker string, divUSD float64, divPLN float64) {
	data := SlackPost{
		Ticker: ticker,
		DivUSD: divUSD,
		DivPLN: divPLN,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "http://localhost:9009/", body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
