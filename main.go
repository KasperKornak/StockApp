package main

import (
	"net/http"

	"github.com/KasperKornak/StockApp/models"
	"github.com/KasperKornak/StockApp/routes"
	"github.com/KasperKornak/StockApp/utils"
)

func main() {
	models.Init()
	utils.LoadTemplates("templates/*.html")
	r := routes.NewRouter()
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
