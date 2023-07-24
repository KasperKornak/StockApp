package main

import (
	"log"
	"net/http"

	"github.com/KasperKornak/StockApp/models"
	"github.com/KasperKornak/StockApp/routes"
	"github.com/KasperKornak/StockApp/utils"
	"github.com/robfig/cron/v3"
)

func main() {
	models.Init()
	utils.LoadTemplates("templates/*.html")
	r := routes.NewRouter()
	http.Handle("/", r)
	c := cron.New()
	c.AddFunc("00 16 * * *", func() {
		log.Println("started")
		models.CalculateDividends()
		models.UpdateStockDb()
		models.UpdateUserList()
		log.Println("ended")
	})

	c.Start()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
