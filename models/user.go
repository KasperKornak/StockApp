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

// handle various cases of errors during login/registering
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidLogin  = errors.New("invalid login")
	ErrUsernameTaken = errors.New("username taken")
)

// to id the user in redis
type User struct {
	Id int64
}

// used to create initial year summary document
type initYearSummary struct {
	Year         int     `bson:"year"`
	DividendsYTD float64 `bson:"dividendytd"`
	DividendTax  float64 `bson:"dividendtax"`
	Ticker       string  `bson:"ticker"`
}

// no idea what it does
// TODO: check if necessary
type MonthData struct {
	Jan float64 `json:"Jan"`
	Feb float64 `json:"Feb"`
	Mar float64 `json:"Mar"`
	Apr float64 `json:"Apr"`
	May float64 `json:"May"`
	Jun float64 `json:"Jun"`
	Jul float64 `json:"Jul"`
	Aug float64 `json:"Aug"`
	Sep float64 `json:"Sep"`
	Oct float64 `json:"Oct"`
	Nov float64 `json:"Nov"`
	Dec float64 `json:"Dec"`
}

// used to create initial month data document
type InitMonthData struct {
	Name  string  `bson:"name"`
	Value float64 `bson:"value"`
}

// used to create initial month data document
type InitMongoMonths struct {
	Year   int             `bson:"year"`
	Ticker string          `bson:"ticker"`
	Months []InitMonthData `bson:"months"`
}

// struct which stores all info about position in user's collection
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
	DivPaid          int     `json:"divpaid" bson:"divpaid"`
	ExDivDate        int     `json:"exdivdate" bson:"exdivdate"`
}

// to aggregate PositionData struct
type Positions struct {
	Ticker string         `json:"ticker" bson:"ticker"`
	Stocks []PositionData `json:"stocks" bson:"stocks"`
}

// used to create a new user record in redis
func NewUser(username string, hash []byte) (*User, error) {
	// check if username is available
	exists, _ := client.HExists("user:by-username", username).Result()
	if exists {
		return nil, ErrUsernameTaken
	}
	id, err := client.Incr("user:next-id").Result()
	if err != nil {
		return nil, err
	}

	// insert a new user into redis
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

// returns the id of user
func (user *User) GetId() (int64, error) {
	return user.Id, nil
}

// returns username based on the id of the user
func (user *User) GetUsername() (string, error) {
	key := fmt.Sprintf("user:%d", user.Id)
	return client.HGet(key, "username").Result()
}

// returns hash of the user
func (user *User) GetHash() ([]byte, error) {
	key := fmt.Sprintf("user:%d", user.Id)
	return client.HGet(key, "hash").Bytes()
}

// compares provided password with hash and authenticates
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

// literally what it says
func GetUserById(id int64) (*User, error) {
	return &User{id}, nil
}

// literally what it says
func GetUserByUsername(username string) (*User, error) {
	id, err := client.HGet("user:by-username", username).Int64()
	if err == redis.Nil {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return GetUserById(id)
}

// authenticate the user
func AuthenticateUser(username, password string) (*User, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, user.Authenticate(password)
}

// generates password hash, inserts initial crucial documents into user's collection in mongodb
func RegisterUser(username, password string) error {
	cost := bcrypt.DefaultCost
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return err
	}
	// make place for new user in redis
	_, err = NewUser(username, hash)

	// create a new collection for user
	collection := MongoClient.Database("users").Collection(username)

	// create documents and insert them into collection
	yearSummary := &initYearSummary{
		Year:         time.Now().Year(),
		DividendsYTD: 0.0,
		DividendTax:  0.0,
		Ticker:       "YEAR_SUMMARY",
	}

	collection.InsertOne(context.TODO(), yearSummary)

	months := []InitMonthData{
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

	doc := InitMongoMonths{
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

	// add username to username list
	utils := MongoClient.Database("users").Collection("stockUtils")
	userListFilter := bson.M{"ticker": "ALL_USERNAMES"}
	var userList UsernamesDocument
	_ = utils.FindOne(context.TODO(), userListFilter).Decode(&userList)
	if err != nil {
		log.Println("func RegisterUser: ", err)
	}

	userList.Usernames = append(userList.Usernames, username)
	_, err = utils.UpdateOne(context.TODO(), userListFilter, bson.M{"$set": bson.M{"usernames": userList.Usernames}})
	log.Println(userList.Usernames)

	return err
}

// no idea what it does
// TODO: check if necessary
func GetName(r *http.Request) int64 {
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, _ := untypedUserId.(int64)
	return userId
}

// after deleting a position transfer YTD dividend data to DELETED_SUM document
func TransferDivs(username string, ticker string) {
	// init variables and open connection to mongodb
	stocks := MongoClient.Database("users").Collection(username)
	deletedStocks := ModelGetStockByTicker("DELETED_SUM", username)

	toDelete := ModelGetStockByTicker(ticker, username)
	deletedStocks.DivYTD = deletedStocks.DivYTD + toDelete.DivYTD
	deletedStocks.DivPLN = deletedStocks.DivPLN + toDelete.DivPLN

	// set update
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

	// apply the update
	updateOptions := options.Update().SetArrayFilters(arrayFilters)
	filter := bson.M{"stocks": bson.M{"$elemMatch": bson.M{"ticker": ticker}}, "ticker": "positions"}
	_, err := stocks.UpdateOne(context.TODO(), filter, update, updateOptions)
	if err != nil {
		log.Println("func RegisterUser:; Error updating document:", err)
	}
}

// returns position data of given user, provided ticker and username
func ModelGetStockByTicker(ticker string, username string) PositionData {
	// open connection to mongodb and init some variables
	stocks := MongoClient.Database("users").Collection(username)
	var stockData struct {
		Stocks []PositionData `bson:"stocks"`
	}

	filter := bson.M{"ticker": "positions", "stocks.ticker": ticker}
	err := stocks.FindOne(context.TODO(), filter).Decode(&stockData)
	if err != nil {
		log.Println("func ModelGetStockByTicker", err)
		return PositionData{}
	}

	// iterate over retrieved stocks, if ticker checks out - return the position data
	for _, stock := range stockData.Stocks {
		if stock.Ticker == ticker {
			return stock
		}
	}

	return PositionData{}
}

// used to handle request to edit positions
func EditPosition(edit PositionData, username string) PositionData {
	// get the current state of user's position data
	currState := ModelGetStockByTicker(edit.Ticker, username)

	// init PositionData variable and check which fields user has changed
	finalVersion := PositionData{}
	finalVersion.Ticker = edit.Ticker

	// if user want to edit DELETED_SUM document, it is only possible to edit its divpln and divytd fields
	if edit.Ticker == "DELETED_SUM" {
		finalVersion.DivPLN = edit.DivPLN
		finalVersion.DivYTD = edit.DivYTD
	} else {
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
		finalVersion.DivPaid = currState.DivPaid
		finalVersion.Currency = currState.Currency
		finalVersion.DivQuarterlyRate = currState.DivQuarterlyRate
		finalVersion.NextPayment = currState.NextPayment
		finalVersion.PrevPayment = currState.PrevPayment
		finalVersion.SharesAtExDiv = currState.SharesAtExDiv
		finalVersion.ExDivDate = currState.ExDivDate
	}
	return finalVersion
}

// used to update divpaid field
func UpdateExDivDate() {
	// init variables, retrieve username list and all tracked stocks
	currUserList := RetrieveUsers()
	currentTime := int(time.Now().Unix())

	// iterate over username list, connect to their collection
	for _, username := range currUserList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions
		curUserFilter := bson.M{"ticker": "positions"}
		err := currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		if err != nil {
			log.Println("func UpdateExDivDate:", err)
		}

		// iterate over positions of user and check if dividend has been payed out
		for _, position := range currUserStocks.Stocks {
			if position.NextPayment <= currentTime {
				position.DivPaid = 1
				// to delete in near future
				log.Println("func UpdateExDivDate: updated divPaid for user: ", username, "\nposition: ", position.Ticker)
			} else {
				position.DivPaid = 0
			}
		}
		// update user's positions
		update := bson.M{
			"$set": bson.M{
				"stocks": currUserStocks.Stocks,
			},
		}
		_, err = currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println("func UpdateExDivDate: ", err)
		}
	}
}

// used to check shares at exdiv
func CheckSharesAtExdiv() {
	// init variables, retrieve username list and all tracked stocks
	currUserList := RetrieveUsers()
	currentTime := int(time.Now().Unix())

	// iterate over username list, connect to their collection
	for _, username := range currUserList.Usernames {
		currUserCollection := MongoClient.Database("users").Collection(username)
		var currUserStocks Positions
		curUserFilter := bson.M{"ticker": "positions"}
		err := currUserCollection.FindOne(context.TODO(), curUserFilter).Decode(&currUserStocks)
		if err != nil {
			log.Println("func CheckSharesAtExdiv:", err)
		}

		// iterate over positions of user and check if dividend has been payed out
		for _, position := range currUserStocks.Stocks {
			if Abs(position.ExDivDate-currentTime) <= 48*60*60 {
				position.SharesAtExDiv = position.Shares
			}
		}
		// update user's positions
		update := bson.M{
			"$set": bson.M{
				"stocks": currUserStocks.Stocks,
			},
		}
		_, err = currUserCollection.UpdateOne(context.TODO(), curUserFilter, update)
		if err != nil {
			log.Println("func CheckSharesAtExdiv: ", err)
		}
	}
}
