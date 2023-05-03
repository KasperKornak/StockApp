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
curl -X POST -H "Content-Type: application/json" -d '{"ticker": "PG", "shares": 6, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.9407, "divytd": 0.0, "divpln": 0.0, "nextpayment": 1684101600, "prevpayment":1676415600}' http://localhost:9010/create
```
