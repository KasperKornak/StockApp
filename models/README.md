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
