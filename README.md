# StockApp
## Used paths
| URL | Method | JSON | Description |
| --- | --- | --- | --- |
|`localhost:9010/stocks` | **GET** | - | Used to get full stock list|
|`localhost:9010/stock/{ticker}` | **GET** | - | Used to get one company |
|`localhost:9010/delete` | **DELETE** | `{"symbol": "RIO.L"}` | Delete the whole position |
|`localhost:9010/create` | **POST** |`curl -X POST -H "Content-Type: application/json" -d '{"ticker": "PG", "shares": 6, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.9407, "divytd": 0.0, "divpln": 0.0, "nextpayment": 1684101600, "prevpayment":1676415600}' http://localhost:9010/create` | Creates new position|
|`localhost:9010/update` | **PUT** |`curl -X POST -H "Content-Type: application/json" -d '{"ticker": "PG", "shares": 6, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.9407, "divytd": 0.0, "divpln": 0.0, "nextpayment": 1684101600, "prevpayment":1676415600}' http://localhost:9010/create` | Update exisiting position|

## Functions used
| Function | Package | Description |
|   ---    |    ---    |    ---      |
| `init()` | **main** | Checks if deleted stocks and summary documents exists, if not creates them. Updates summary document at the start of the StockApp.  |
| `main()` | **main** | Responsible for running app on port 9010 and cron job which synchronizes dividend data. |
|  `MongoConnect()` | **config** | Retrieves credentials, connects to MongoDB. Returns client of type `*mongo.Client`. |
|`GetPaymentDate()`|**controllers**|Used to update dividend dates for the company, uses UNIX timestamps.|
|`CheckPayment()`|**controllers**|Check if payment has been received, updates suitable fields in MongoDB.|
|`UpdateSummary()`| **controllers** |Updates *YEAR_SUMMARY* and *DELETED_SUM* documents in MongoDB.|
|`CheckYear()`| **controllers** |If current year isn't the same as in documents, updates dividend-related fields, prepares *DELETED_SUM*, *YEAR_SUMMARY* and rest of the documents in Mongo for calculations in new year.|
|`GetStocks(w http.ResponseWriter, r *http.Request)`|**controllers**|Used for handling requests made to */stocks* endpoint.|
|`GetStockByTicker(w http.ResponseWriter, r *http.Request)`|**controllers** |Used for handling requests made to */stock/{ticker}* endpoint. |
|`DeletePosition(w http.ResponseWriter, r *http.Request)`|**controllers**|Used for handling requests made to */delete* endpoint.|
|`CreatePosition(w http.ResponseWriter, r *http.Request)`|**controllers**|Used for handling requests made to */create* endpoint.|
|`UpdatePosition(w http.ResponseWriter, r *http.Request)`|**controllers**| Used for handling requests made to */update* endpoint.|
|`StocksHTML(w http.ResponseWriter, r *http.Request)`|**controllers**| Used to render dividend table on */home*.|
|`MongoMiddleware(client *mongo.Client)`|**middleware**|Used to share Mongo client between multiple functions.|
|`func ModelGetStocks(Client *mongo.Client)`|**models**|Extension of *GetStocks* - extracts all stocks from MongoDB in form of type *Company* slice.|
|`func ModelGetStockByTicker(ticker string, Client *mongo.Client)`|**models**|Extension of *GetStockByTicker* - extract single company from MongoDB, returns *Company* struct.|
|`func ModelDeletePosition(ticker string, Client *mongo.Client)`|**models**|Extension of *DeletePosition* sends query to MongoDB to delete company record.|
|`func ModelCreatePosition(ticker string, shares int, domestictax int, currency string, divQuarterlyRate float64, divytd float64, divpln float64, nextpayment int, prevpayment int, Client *mongo.Client)`|**models**|Extension of *CreatePosition* sends position data to MongoDB.|
|`func ModelUpdatePosition(ticker string, shares int, domestictax int, currency string, divQuarterlyRate float64, divytd float64, divpln float64, nextpayment int, prevpayment int, Client *mongo.Client)`|**models**|Extension of *UpdatePosition*, updates position in MongoDB.|
|`TransferDivs(divRec float64, divTax float64, Client *mongo.Client)`|**models**|Component of *DeletePosition* - transfers tax and dividend data to *DELETED_SUM* document.|


## Links
[Folder layout](https://www.youtube.com/watch?v=Y7kuW1qyDng)

[Mongo](https://www.mongodb.com)

[Mongo connection driver](https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection/#connection-guide)

[Finance Go](https://piquette.io/projects/finance-go/)

[Cron jobs in Go](https://www.airplane.dev/blog/creating-golang-cron-jobs)

[mongo on k8s](https://www.mongodb.com/kubernetes)

[mongo go example](https://www.youtube.com/watch?v=D3jhplPWqnA) 

