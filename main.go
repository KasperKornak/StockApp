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
	// init connections to dbs, load templates and start router
	models.Init()
	utils.LoadTemplates("templates/*.html")
	r := routes.NewRouter()
	http.Handle("/", r)

	// cronjob which updates dbs
	c := cron.New()
	c.AddFunc("10 40 * * *", func() {
		// log.Println("started")
		// models.CalculateDividends()
		// log.Println("calculating divs finished")
		// models.UpdateStockDb()
		// log.Println("updating stock db finished")
		models.UpdateUserList()
		log.Println("updating user list finished")
		// models.UpdateExDivDate()
		// log.Println("ad hoc fix of divPaid variable fixed")
		// log.Println("ended")
	})
	c.Start()

	// start serving on port 80, change to 443 n future
	log.Fatal(http.ListenAndServe(":80", nil))
}
