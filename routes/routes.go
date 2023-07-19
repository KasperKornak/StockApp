package routes

import (
	"context"
	"encoding/json"

	"log"
	"net/http"
	"time"

	"github.com/KasperKornak/StockApp/middleware"
	"github.com/KasperKornak/StockApp/models"
	"github.com/KasperKornak/StockApp/sessions"
	"github.com/KasperKornak/StockApp/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
)

type WebUser struct {
	Id           int64
	Username     string
	DividendsYTD float64            `json:"dividendytd" bson:"dividendytd"`
	DividendTax  float64            `json:"dividendtax" bson:"dividendtax"`
	Months       map[string]float64 `json:"months" bson:"months"`
	Bar          *charts.Bar
}

type MongoSummary struct {
	ID           string      `bson:"_id"`
	Year         int         `bson:"year"`
	DividendsYTD float64     `bson:"dividendytd"`
	DividendTax  float64     `bson:"dividendtax"`
	Ticker       string      `bson:"ticker"`
	Months       []MonthData `bson:"months"`
}

type MonthData struct {
	Name  string  `bson:"name"`
	Value float64 `bson:"value"`
}

type MongoMonths struct {
	ID     string      `bson:"_id"`
	Year   int         `bson:"year"`
	Ticker string      `bson:"ticker"`
	Months []MonthData `bson:"months"`
}

type DeletePosition struct {
	Ticker string `json:"ticker" bson:"ticker"`
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", middleware.AuthRequired(indexGetHandler)).Methods("GET")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/logout", logoutGetHandler).Methods("GET")
	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/positions", middleware.AuthRequired(positionsGetHandler)).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")
	r.HandleFunc("/logout", registerPostHandler).Methods("GET")
	r.HandleFunc("/api/data", barDataHandler).Methods("GET")
	r.HandleFunc("/api/positions", positionsDataHandler).Methods("GET")
	r.HandleFunc("/api/update", updatePutHandler).Methods("PUT")
	r.HandleFunc("/api/update", positionsDataHandler).Methods("POST")
	r.HandleFunc("/api/update", updateDeleteHandler).Methods("DELETE")
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	return r
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	user, err := models.AuthenticateUser(username, password)
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
	session, _ := sessions.Store.Get(r, "session")
	session.Values["user_id"] = userId
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessions.Store.Get(r, "session")
	delete(session.Values, "user_id")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	err := models.RegisterUser(username, password)
	if err == models.ErrUsernameTaken {
		utils.ExecuteTemplate(w, "register.html", "username taken")
		return
	} else if err != nil {
		utils.InternalServerError(w)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	var CurrUser WebUser
	var tempUser models.User
	var mongotransfer MongoSummary

	CurrUser.Id = models.GetName(r)
	tempUser.Id = CurrUser.Id
	CurrUser.Username, _ = tempUser.GetUsername()
	stocks := models.MongoClient.Database("users").Collection(CurrUser.Username)
	filter := bson.M{"ticker": "YEAR_SUMMARY", "year": time.Now().Year()}

	mongotransfer = MongoSummary{}
	err := stocks.FindOne(context.TODO(), filter).Decode(&mongotransfer)

	if err != nil {
		log.Println(err)
	}
	CurrUser.DividendTax = mongotransfer.DividendTax
	CurrUser.DividendsYTD = mongotransfer.DividendsYTD

	utils.ExecuteTemplate(w, "index.html", CurrUser)
}

func barDataHandler(w http.ResponseWriter, r *http.Request) {
	var CurrUser WebUser
	var tempUser models.User
	var mongomonths MongoMonths

	CurrUser.Id = models.GetName(r)
	tempUser.Id = CurrUser.Id
	CurrUser.Username, _ = tempUser.GetUsername()
	stocks := models.MongoClient.Database("users").Collection(CurrUser.Username)
	filter := bson.M{"ticker": "MONTH_SUMARY", "year": time.Now().Year()}

	var mongotransfer MongoSummary
	err := stocks.FindOne(context.TODO(), filter).Decode(&mongotransfer)
	if err != nil {
		log.Println(err)
	}

	// Assign the fetched data to the MongoMonths struct
	mongomonths = MongoMonths{
		ID:     mongotransfer.ID,
		Year:   mongotransfer.Year,
		Ticker: mongotransfer.Ticker,
		Months: mongotransfer.Months,
	}

	// Marshal the chart data to JSON and send as response
	jsonData, err := json.Marshal(mongomonths)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func positionsDataHandler(w http.ResponseWriter, r *http.Request) {
	var toTable models.Positions
	var tempUser models.User
	id := models.GetName(r)
	tempUser.Id = id
	username, _ := tempUser.GetUsername()
	stocks := models.MongoClient.Database("users").Collection(username)
	filter := bson.M{"ticker": "positions"}

	err := stocks.FindOne(context.TODO(), filter).Decode(&toTable)
	if err != nil {
		log.Println(err)
	}

	jsonData, err := json.Marshal(toTable)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func positionsGetHandler(w http.ResponseWriter, r *http.Request) {
	utils.ExecuteTemplate(w, "positions.html", nil)
}

func updatePutHandler(w http.ResponseWriter, r *http.Request) {

}

func updateDeleteHandler(w http.ResponseWriter, r *http.Request) {
	var toDelete DeletePosition
	var tempUser models.User
	id := models.GetName(r)
	tempUser.Id = id
	username, _ := tempUser.GetUsername()
	err := json.NewDecoder(r.Body).Decode(&toDelete)
	if err != nil {
		log.Println(err)
	}

	models.TransferDivs(username, toDelete.Ticker)

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
		log.Print("contact administrator something went wrong :| \n")
	}
}
