package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KasperKornak/StockApp/sessions"
	"github.com/go-redis/redis"
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
	Ticker           string  `json:"ticker"`
	Shares           int     `json:"shares"`
	Domestictax      int     `json:"domestictax"`
	Currency         string  `json:"currency"`
	DivQuarterlyRate float64 `json:"divquarterlyrate" bson:"divquarterlyrate"`
	DivYTD           float64 `json:"divytd" bson:"divytd"`
	DivPLN           float64 `json:"divpln" bson:"divpln"`
	NextPayment      int     `json:"nextpayment" bson:"nextpayment"`
	PrevPayment      int     `json:"prevpayment" bson:"prevpayment"`
}

type Positions struct {
	Ticker string         `json:"ticker" bson:"year"`
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
	return err
}

func GetName(r *http.Request) int64 {
	session, _ := sessions.Store.Get(r, "session")
	untypedUserId := session.Values["user_id"]
	userId, _ := untypedUserId.(int64)
	return userId
}
