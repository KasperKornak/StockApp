# StockApp
## Used paths
| URL | Method | JSON | Description |
| --- | --- | --- | --- |
|`localhost:9010/stocks` | **GET** | - | Used to get full stock list|
|`localhost:9010/stock/{ticker}` | **GET** | - | Used to get one company |
|`localhost:9010/delete` | **DELETE** | `{"symbol": "RIO.L"}` | Delete the whole position |
|`localhost:9010/create` | **POST** |`{"ticker": "MSFT", "shares": 2, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.68, "divytd": 14.2, "divpln": 67.32, "nextpayment": "2023-06-01"}` | Creates new position|
|`localhost:9010/update` | **PUT** |`{"ticker": "MSFT", "shares": 2, "domestictax": 4, "currency": "USD", "divquarterlyrate": 0.68, "divytd": 14.2, "divpln": 67.32, "nextpayment": "2023-06-01"}` | Update exisiting position|

## TODO list
- [ ] add better error handling,
- [x] configure years collection to work with `/create` path,
- [x] add ".L" stock support to getting one company **works with requests**,
- [x] add simple HTML table view,

## Links
[Folder layout](https://www.youtube.com/watch?v=Y7kuW1qyDng)

[Mongo](https://www.mongodb.com)

[Mongo connection driver](https://www.mongodb.com/docs/drivers/go/current/fundamentals/connection/#connection-guide)

[Finance Go](https://piquette.io/projects/finance-go/)

[Cron jobs in Go](https://www.airplane.dev/blog/creating-golang-cron-jobs)

[mongo on k8s](https://www.mongodb.com/kubernetes)

[mongo go example](https://www.youtube.com/watch?v=D3jhplPWqnA) 

