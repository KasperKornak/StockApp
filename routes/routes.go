package routes

import (
	"context"
	"encoding/json"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KasperKornak/StockApp/middleware"
	"github.com/KasperKornak/StockApp/models"
	"github.com/KasperKornak/StockApp/sessions"
	"github.com/KasperKornak/StockApp/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
)

// used to render username, months, bar chart and dividend data on home page
type WebUser struct {
	Id           int64
	Username     string
	DividendsYTD float64            `json:"dividendytd" bson:"dividendytd"`
	DividendTax  float64            `json:"dividendtax" bson:"dividendtax"`
	Months       map[string]float64 `json:"months" bson:"months"`
	Bar          *charts.Bar
}

// struct used to retrieve MONGO_MONTHS document from mongodb
type MongoSummary struct {
	ID           string      `bson:"_id"`
	Year         int         `bson:"year"`
	DividendsYTD float64     `bson:"dividendytd"`
	DividendTax  float64     `bson:"dividendtax"`
	Ticker       string      `bson:"ticker"`
	Months       []MonthData `bson:"months"`
}

// used to store each individual month
type MonthData struct {
	Name  string  `bson:"name"`
	Value float64 `bson:"value"`
}

// used to store all months
type MongoMonths struct {
	ID     string      `bson:"_id"`
	Year   int         `bson:"year"`
	Ticker string      `bson:"ticker"`
	Months []MonthData `bson:"months"`
}

// used to handle delete position requests
type DeletePosition struct {
	Ticker string `json:"ticker" bson:"ticker"`
}

// router used to handle all user and api endpoints
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// user endpoints
	r.HandleFunc("/", middleware.AuthRequired(indexGetHandler)).Methods("GET")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/positions", middleware.AuthRequired(positionsGetHandler)).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")
	r.HandleFunc("/docs", tutorialHandler).Methods("GET")

	// api endpoints
	r.HandleFunc("/api/data", middleware.AuthRequired(barDataHandler)).Methods("GET")
	r.HandleFunc("/api/positions", middleware.AuthRequired(positionsDataHandler)).Methods("GET")
	r.HandleFunc("/api/update", middleware.AuthRequired(updateEditHandler)).Methods("PUT")
	r.HandleFunc("/api/update", middleware.AuthRequired(updateAddHandler)).Methods("POST")
	r.HandleFunc("/api/update", middleware.AuthRequired(updateDeleteHandler)).Methods("DELETE")
	r.HandleFunc("/api/month", middleware.AuthRequired(monthSummaryUpdateHandler)).Methods("POST")

	// file server for all static files on website
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	return r
}

// execute template of login page
func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "login.html", nil)
}

// handle login request
func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	user, err := models.AuthenticateUser(username, password)

	// throw an error if auth unsuccessful
	if err != nil {
		switch err {
		case models.ErrUserNotFound:
			utils.ExecuteTemplate(w, "login.html", "unknown user")
		case models.ErrInvalidLogin:
			utils.ExecuteTemplate(w, "login.html", "invalid login")
		default:
			utils.InternalServerError(w)
		}
		return
	}
	userId, err := user.GetId()
	if err != nil {
		utils.InternalServerError(w)
		return
	}

	// get a session if successful
	session, err := sessions.Store.Get(r, "session")
	if err != nil {
		log.Println("func loginPostHandler: ", err)
	}
	session.Values["user_id"] = userId
	session.Save(r, w)

	// redirect to homepage
	http.Redirect(w, r, "/", http.StatusFound)
}

// logout
func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	// delete user's session
	session, err := sessions.Store.Get(r, "session")
	if err != nil {
		log.Println("func logoutGetHandler: ", err)
	}
	delete(session.Values, "user_id")
	session.Save(r, w)

	// redirect to login page
	http.Redirect(w, r, "/login", http.StatusFound)
}

// execute tempalate of register website
func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "register.html", nil)
}

// handle register request
func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	// extract register data
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	// register user if unsuccessful throw an error
	err := models.RegisterUser(username, password)
	if err == models.ErrUsernameTaken {
		utils.ExecuteTemplate(w, "register.html", "username taken")
		return
	} else if err != nil {
		utils.InternalServerError(w)
		return
	}

	// redirect to login page
	http.Redirect(w, r, "/login", http.StatusFound)
}

// render home page
func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	// init variable handling user data and bar data
	var CurrUser WebUser
	var tempUser models.User
	var mongotransfer MongoSummary
	var err error
	CurrUser.Id = models.GetName(r)
	tempUser.Id = CurrUser.Id
	CurrUser.Username, err = tempUser.GetUsername()
	if err != nil {
		log.Println("func indexGetHandler: ", err)
	}

	// select correct collection in mongodb and retrieve YEAR_SUMMARY data
	stocks := models.MongoClient.Database("users").Collection(CurrUser.Username)
	filter := bson.M{"ticker": "YEAR_SUMMARY", "year": time.Now().Year()}

	mongotransfer = MongoSummary{}
	err = stocks.FindOne(context.TODO(), filter).Decode(&mongotransfer)

	if err != nil {
		log.Println("func indexGetHandler", err)
	}

	// final data to be rendered
	CurrUser.DividendTax = mongotransfer.DividendTax
	CurrUser.DividendsYTD = mongotransfer.DividendsYTD
	utils.ExecuteTemplate(w, "index.html", CurrUser)
}

// used to handle requests for rendering each month's data
func barDataHandler(w http.ResponseWriter, r *http.Request) {
	// init variables to handle rendering template
	var CurrUser WebUser
	var tempUser models.User
	var mongomonths MongoMonths
	var err error
	CurrUser.Id = models.GetName(r)
	tempUser.Id = CurrUser.Id
	CurrUser.Username, err = tempUser.GetUsername()
	if err != nil {
		log.Println("func barDataHandler: ", err)
	}
	// retrieve MONTH_SUMMARY from mongodb
	stocks := models.MongoClient.Database("users").Collection(CurrUser.Username)
	filter := bson.M{"ticker": "MONTH_SUMARY", "year": time.Now().Year()}
	var mongotransfer MongoSummary
	err = stocks.FindOne(context.TODO(), filter).Decode(&mongotransfer)
	if err != nil {
		log.Println("func barDataHandler: ", err)
	}

	// assign the fetched data to the MongoMonths struct
	mongomonths = MongoMonths{
		ID:     mongotransfer.ID,
		Year:   mongotransfer.Year,
		Ticker: mongotransfer.Ticker,
		Months: mongotransfer.Months,
	}

	// marshal the chart data to JSON and send as response
	jsonData, err := json.Marshal(mongomonths)
	if err != nil {
		log.Println("func barDataHandler: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// handle data needed for /positions
func positionsDataHandler(w http.ResponseWriter, r *http.Request) {
	// init variables needed to handle data
	var toTable models.Positions
	var tempUser models.User
	var err error
	id := models.GetName(r)
	tempUser.Id = id
	username, err := tempUser.GetUsername()
	if err != nil {
		log.Println("func positionsDataHandler: ", err)
	}
	// retrieve positions data
	stocks := models.MongoClient.Database("users").Collection(username)
	filter := bson.M{"ticker": "positions"}
	err = stocks.FindOne(context.TODO(), filter).Decode(&toTable)
	if err != nil {
		log.Println("func positionsDataHandler: ", err)
	}

	// marshal data and send it back
	jsonData, err := json.Marshal(toTable)
	if err != nil {
		log.Println("func positionsDataHandler: ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// render positions page
func positionsGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "positions.html", nil)
}

// handler for editing position data
func updateEditHandler(w http.ResponseWriter, r *http.Request) {
	// init variable to decode received json data
	var toEdit models.PositionData
	err := json.NewDecoder(r.Body).Decode(&toEdit)
	if err != nil {
		log.Println("func updateEditHandler: ", err)
	}

	// user check
	var tempUser models.User
	id := models.GetName(r)
	tempUser.Id = id
	username, err := tempUser.GetUsername()
	if err != nil {
		log.Println("func updateEditHandler: ", err)
	}

	// make ticker and currency to upper to search mongodb
	toEdit.Currency = strings.ToUpper(toEdit.Currency)
	toEdit.Ticker = strings.ToUpper(toEdit.Ticker)

	// edit position and update it in mongodb
	edited := models.EditPosition(toEdit, username)
	stocks := models.MongoClient.Database("users").Collection(username)
	filter := bson.M{"stocks.ticker": toEdit.Ticker}
	update := bson.M{"$set": bson.M{"stocks.$": edited}}
	_, err = stocks.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println("func updateEditHandler; Error updating document:", err)
	}

	// update YEAR_SUMMARY document
	models.UpdateSummary(username)
}

// handler used to delete position
func updateDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// init variables
	var toDelete DeletePosition
	var tempUser models.User
	id := models.GetName(r)
	tempUser.Id = id
	username, err := tempUser.GetUsername()
	if err != nil {
		log.Println("func updateDeleteHandler: ", err)
	}

	// decode received json data
	err = json.NewDecoder(r.Body).Decode(&toDelete)
	if err != nil {
		log.Println("func updateDeleteHandler: ", err)
	}

	// check if user wants to delte DELETED_SUM document, if true return
	if toDelete.Ticker == "DELETED_SUM" {
		log.Println("can't delete this position")
		return
	}

	// after deleting position transfer dividends received
	models.TransferDivs(username, toDelete.Ticker)

	// delete document from mongodb
	toDelete.Ticker = strings.ToUpper(toDelete.Ticker)
	stocks := models.MongoClient.Database("users").Collection(username)
	filter := bson.M{
		"ticker": "positions",
	}
	update := bson.M{
		"$pull": bson.M{
			"stocks": bson.M{
				"ticker": toDelete.Ticker,
			},
		},
	}
	result, err := stocks.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("error deleting document: %v\n", err)
	}

	if result.ModifiedCount != 1 {
		log.Print("document wasn't deleted, because it doesn't exist \n")
	}
}

// add position handler
func updateAddHandler(w http.ResponseWriter, r *http.Request) {
	// init variable to decode received data
	var toAdd models.PositionData
	err := json.NewDecoder(r.Body).Decode(&toAdd)
	if err != nil {
		log.Println("func updateAddHandler: ", err)
	}

	// check if user wants to add DELETED_SUM, if yes return
	if toAdd.Ticker == "DELETED_SUM" {
		log.Println("can't add this position")
		return
	}

	// make ticker uppercase to handle availabilty func
	toAdd.Ticker = strings.ToUpper(toAdd.Ticker)

	// check if ticker is available in stockUtils
	isTickerAvailable := models.CheckTickerAvailabilty(toAdd.Ticker)

	// init SharesAtExDiv
	toAdd.SharesAtExDiv = toAdd.Shares
	if !isTickerAvailable {
		log.Println("ticker unavailable!")
		return
	}

	// insert document to users collection in mongodb
	var tempUser models.User
	id := models.GetName(r)
	tempUser.Id = id
	username, err := tempUser.GetUsername()
	if err != nil {
		log.Println("func updateAddHandler: ", err)
	}
	stocks := models.MongoClient.Database("users").Collection(username)
	filter := bson.M{
		"ticker": "positions",
	}
	update := bson.M{"$push": bson.M{"stocks": toAdd}}
	_, err = stocks.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println("func updateAddHandler;Error updating document: ", err)
	}

	// update summary of user and fetch latest timestamps
	models.UpdateSummary(username)
	models.GetTimestamps(toAdd.Ticker, username)
}

// used to handle editing month data at home page
func monthSummaryUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// init variables
	var tempUser models.User
	var currMonth, editedMonthValues models.InitMongoMonths
	id := models.GetName(r)
	tempUser.Id = id
	username, err := tempUser.GetUsername()
	if err != nil {
		log.Println("func monthSummaryUpdateHandler: ", err)
	}

	// get the current month data
	stocks := models.MongoClient.Database("users").Collection(username)
	monthFilter := bson.M{"ticker": "MONTH_SUMARY", "year": time.Now().Year()}
	err = stocks.FindOne(context.TODO(), monthFilter).Decode(&currMonth)
	if err != nil {
		log.Println("func monthSummaryUpdateHandler: ", err)
	}

	// decode month update data
	err = json.NewDecoder(r.Body).Decode(&editedMonthValues)
	if err != nil {
		log.Println("func monthSummaryUpdateHandler: ", err)
	}

	// update month data in mongodb
	updateDoc := bson.M{"$set": editedMonthValues}
	_, err = stocks.UpdateOne(context.TODO(), monthFilter, updateDoc)
	if err != nil {
		log.Println("func monthSummaryUpdateHandler: ", err)
	}
}

// execute /docs template
func tutorialHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "docs.html", nil)
}
