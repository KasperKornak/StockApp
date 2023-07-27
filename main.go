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
	c.AddFunc("18 23 * * *", func() {
		log.Println("started")
		models.CalculateDividends()
		log.Println("calculating divs finished")
		models.UpdateStockDb()
		log.Println("updating stock db finished")
		models.UpdateUserList()
		log.Println("updating user list finished")
		log.Println("ended")
	})

	c.Start()

	log.Fatal(http.ListenAndServe(":80", nil))
}
