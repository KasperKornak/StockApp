mongosh:
```bash
# setup
use stock
db.createCollection("tickers")
db.createCollection("years")
db.createUser({user: "stock", pwd: "unlock", roles :["dbAdmin"]})

# verify stock creation
db.tickers.find()

# delete document
db.tickers.remove({_id: ObjectId("644c0a2b4da3d982756b4e5d")})
```

terminal:
```bash
cd cmd
go run main.go
curl -H "Content-Type: application/json" --request POST -d '{"ticker": "MSFT", "shares": 2, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.68, "divytd": 14.2, "divpln": 67.32, "nextpayment": "2023-06-01"}' http://localhost:9010/create
```
