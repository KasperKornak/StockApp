# `models`
Important package of app with the biggest impact on the whole project. Connections between Polygon, as well as databases, are dependent on code from this package. Internal processes are defined here. 

## `db.go`
### functions
|func|Description|
|---|---|
|**Init**|Used to open connections to databases.|
|**registerDisconnectHandler**|Listens to process-kill signal and closes connection to MongoDB.|

## `polygon.go`
### functions
|func|Description|
|---|---|
|**CheckTickerAvailabilty**|Used to check if ticker inputed by user is available on Polygon. Returns boolean type.|
|**AddTickerToDb**|Used to add a new company to database.|
|**GetTimestamps**|Fills in timestamps in users' positions.|
|**PolygonTickerUpdate**|Used to get latest next payment, ex-dividend date and cash amount data of the company. Return next payment and ex-dividend date of type int and cash amount of dividend of type float.|

### stucts
|Struct|Description|
|---|---|
|**PolygonJson**|Used to decode data received from Polygon regarding company dividend information.|
|**PolygonPositionData**|Struct used to store data about individual company in MongoDB collection *stockUtils*.|
|**StockUtils**|Struct used to aggreagate all *PolygonPositionData* structs from MongoDB.|
|**ForexResponse**|Used to decode Forex data response from Polygon.|

## `internalServices.go`
### functions
|func|Description|
|---|---|
|**RetrieveUsers**|Retrieve username list; used to handle user dividend data updates.|
|**UpdateUserList**|Iterates over users and updates their dividend-related data in positions, like NextPayment, ExDivDate, etc.|
|**RetrieveAvailableStocks**|Returns all tracked tickers together with their data from *stockUtils*.|
|**UpdateStockDb**|Used to update dividend-related data in tracked tickers document; should run last in cronjob, bypasses Polygon API restrictions by counter and sleep.|
|**CalculateDividends**|Check if dividend has been paid out, update positions and summaries documents.|
|**GetForex**|Get the latest USD/PLN pair data from Polygon; update in the future to support NBP API.|
|**Abs**|Returns absolute value of a number. Used in calculating shares at ex-div date.|
|**UpdateSummary**|Updates the summary documents of user; called at least once a day.|

### structs
|Struct|Description|
|---|---|
|**UsernameDocument**|Used to handle username list retrieval.|

## `user.go`
### functions
|func|Description|
|---|---|
|**NewUser**|Used to create a new user record in Redis.|
|**GetId**|Returns the Redis ID of the user.|
|**GetUsername**|Returns username based on the ID of the user.|
|**GetHash**|Returns hash of the user.|
|**Authenticate**|Compares provided password with hash and authenticates.|
|**GetUserById**|Get user data by ID.|
|**GetUserByUsername**|Get user data by username.|
|**AuthenticateUser**|Authenticate the user - use **Authenticate** func.|
|**RegisterUser**|Uses **NewUser** func to create a new user in Redis, creates new user's collection and documents in MongoDB.|
|**GetName**|‼️ No idea what it does, check if necessary. ‼️|
|**TransferDivs**|After deleting a position, transfer YTD dividend data to DELETED_SUM document.|
|**ModelGetStockByTicker**|Returns position data of given user, provided ticker and username.|
|**EditPosition**|Used to handle request to edit positions.|

### structs
|Struct|Description|
|---|---|
|**User**|Used to identify user in Redis.|
|**initYearSummary**|Used to create initial year summary document.|
|**MonthData**|‼️ No idea what it does; check if necessary. ‼️|
|**InitMonthData**|Used to create initial month data document.|
|**InitMongoMonths**|Used to create initial month data document; aggregates the upper struct.|
|**PositionData**|Struct which stores all info about position in user's collection.|
|**Positions**|Used to aggregate PositionData struct.|