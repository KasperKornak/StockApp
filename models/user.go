package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KasperKornak/StockApp/sessions"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidLogin  = errors.New("invalid login")
	ErrUsernameTaken = errors.New("username taken")
)

type User struct {
	Id int64
}

type initYearSummary struct {
	Year         int     `bson:"year"`
	DividendsYTD float64 `bson:"dividendytd"`
	DividendTax  float64 `bson:"dividendtax"`
	Ticker       string  `bson:"ticker"`
}

type initMonthData struct {
	Name  string  `bson:"name"`
	Value float64 `bson:"value"`
}

type initMongoMonths struct {
	Year   int             `bson:"year"`
	Ticker string          `bson:"ticker"`
	Months []initMonthData `bson:"months"`
}

type PositionData struct {
	Ticker           string  `json:"ticker" bson:"ticker"`
	Shares           int     `json:"shares" bson:"shares"`
	Domestictax      int     `json:"domestictax" bson:"domestictax"`
	Currency         string  `json:"currency" bson:"currency"`
	DivQuarterlyRate float64 `json:"divquarterlyrate" bson:"divquarterlyrate"`
	DivYTD           float64 `json:"divytd" bson:"divytd"`
	DivPLN           float64 `json:"divpln" bson:"divpln"`
	NextPayment      int     `json:"nextpayment" bson:"nextpayment"`
	PrevPayment      int     `json:"prevpayment" bson:"prevpayment"`
	SharesAtExDiv    int     `json:"sharesatexdiv" bson:"sharesatexdiv"`
}

type Positions struct {
	Ticker string         `json:"ticker" bson:"ticker"`
	Stocks []PositionData `json:"stocks" bson:"stocks"`
}

func NewUser(username string, hash []byte) (*User, error) {
	exists, _ := client.HExists("user:by-username", username).Result()
	if exists {
		return nil, ErrUsernameTaken
	}
	id, err := client.Incr("user:next-id").Result()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("user:%d", id)
	pipe := client.Pipeline()
	pipe.HSet(key, "id", id)
	pipe.HSet(key, "username", username)
	pipe.HSet(key, "hash", hash)
	pipe.HSet("user:by-username", username, id)
	_, err = pipe.Exec()
	if err != nil {
		return nil, err
	}
	return &User{id}, nil
}

func (user *User) GetId() (int64, error) {
	return user.Id, nil
}

func (user *User) GetUsername() (string, error) {
	key := fmt.Sprintf("user:%d", user.Id)
	return client.HGet(key, "username").Result()
}

func (user *User) GetHash() ([]byte, error) {
	key := fmt.Sprintf("user:%d", user.Id)
	return client.HGet(key, "hash").Bytes()
}

func (user *User) Authenticate(password string) error {
	hash, err := user.GetHash()
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrInvalidLogin
	}
	return err
}

func GetUserById(id int64) (*User, error) {
	return &User{id}, nil
}

func GetUserByUsername(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redis.Nil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return GetUserById(id)
}

func AuthenticateUser(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, user.Authenticate(password)
}

func RegisterUser(username, password string) error {
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}
	_, err = NewUser(username, hash)

	collection := MongoClient.Database("users").Collection(username)
	yearSummary := &initYearSummary{
		Year:         time.Now().Year(),
		DividendsYTD: 0.0,
		DividendTax:  0.0,
		Ticker:       "YEAR_SUMMARY",
	}

	collection.InsertOne(context.TODO(), yearSummary)

	// Create the document
	months := []initMonthData{
		{Name: "Jan", Value: 0.0},
		{Name: "Feb", Value: 0.0},
		{Name: "Mar", Value: 0.0},
		{Name: "Apr", Value: 0.0},
		{Name: "May", Value: 0.0},
		{Name: "Jun", Value: 0.0},
		{Name: "Jul", Value: 0.0},
		{Name: "Aug", Value: 0.0},
		{Name: "Sep", Value: 0.0},
		{Name: "Oct", Value: 0.0},
		{Name: "Nov", Value: 0.0},
		{Name: "Dec", Value: 0.0},
	}

	doc := initMongoMonths{
		Year:   time.Now().Year(),
		Ticker: "MONTH_SUMARY",
		Months: months,
	}

	collection.InsertOne(context.TODO(), doc)

	stockDoc := Positions{
		Ticker: "positions",
		Stocks: []PositionData{
			{Ticker: "DELETED_SUM",
				Shares:           0,
				Domestictax:      0,
				Currency:         "USD",
				DivQuarterlyRate: 0.0,
				DivYTD:           0.0,
				DivPLN:           0.0,
				NextPayment:      0,
				PrevPayment:      0},
		},
	}
	collection.InsertOne(context.TODO(), stockDoc)

	return err
}

func GetName(r *http.Request) int64 {
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, _ := untypedUserId.(int64)
	return userId
}

func TransferDivs(username string, ticker string) {
	stocks := MongoClient.Database("users").Collection(username)

	deletedStocks := ModelGetStockByTicker("DELETED_SUM", username)

	toDelete := ModelGetStockByTicker(ticker, username)
	deletedStocks.DivYTD = deletedStocks.DivYTD + toDelete.DivYTD
	deletedStocks.DivPLN = deletedStocks.DivPLN + toDelete.DivPLN

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "stocks.$[elem].divytd", Value: deletedStocks.DivYTD},
			{Key: "stocks.$[elem].divpln", Value: deletedStocks.DivPLN},
		}},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.ticker": "DELETED_SUM"},
		},
	}

	updateOptions := options.Update().SetArrayFilters(arrayFilters)
	filter := bson.M{"stocks": bson.M{"$elemMatch": bson.M{"ticker": ticker}}, "ticker": "positions"}
	_, err := stocks.UpdateOne(context.TODO(), filter, update, updateOptions)
	if err != nil {
		log.Println("Error updating document:", err)
	}
}

func ModelGetStockByTicker(ticker string, username string) PositionData {
	stocks := MongoClient.Database("users").Collection(username)
	var stockData struct {
		Stocks []PositionData `bson:"stocks"`
	}

	filter := bson.M{"stocks": bson.M{"$elemMatch": bson.M{"ticker": ticker}}, "ticker": "positions"}
	err := stocks.FindOne(context.TODO(), filter).Decode(&stockData)
	if err != nil {
		log.Println(err)
		return PositionData{}
	}

	for _, stock := range stockData.Stocks {
		if stock.Ticker == ticker {
			return stock
		}
	}

	return PositionData{}
}

func EditPosition(edit PositionData, username string) PositionData {
	currState := ModelGetStockByTicker(edit.Ticker, username)
	finalVersion := PositionData{}
	finalVersion.Ticker = edit.Ticker
	if edit.Currency != "" {
		finalVersion.Currency = edit.Currency
	} else {
		finalVersion.Currency = currState.Currency
	}

	if edit.Shares != 0 {
		finalVersion.Shares = edit.Shares
	} else {
		finalVersion.Shares = currState.Shares
	}

	if edit.Domestictax != 0 {
		finalVersion.Domestictax = edit.Domestictax
	} else {
		finalVersion.Domestictax = currState.Domestictax
	}

	if edit.DivQuarterlyRate != 0 {
		finalVersion.DivQuarterlyRate = edit.DivQuarterlyRate
	} else {
		finalVersion.DivQuarterlyRate = currState.DivQuarterlyRate
	}

	if edit.DivYTD != 0 {
		finalVersion.DivYTD = edit.DivYTD
	} else {
		finalVersion.DivYTD = currState.DivYTD
	}

	if edit.DivPLN != 0 {
		finalVersion.DivPLN = edit.DivPLN
	} else {
		finalVersion.DivPLN = currState.DivPLN
	}

	if edit.NextPayment != 0 {
		finalVersion.NextPayment = edit.NextPayment
	} else {
		finalVersion.NextPayment = currState.NextPayment
	}

	if edit.PrevPayment != 0 {
		finalVersion.PrevPayment = edit.PrevPayment
	} else {
		finalVersion.PrevPayment = currState.PrevPayment
	}

	return finalVersion
}
