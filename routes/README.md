# `routes`
Comprehensive package of routes and data structures used to handle users' requests.

## User routes
|Route|Handler|Description| Requires auth?|
|---|---|---|---|
|`/`|**indexGetHandler**| Used to render all contents from home page after loggining in. Username, dividend tax and dividends YTD are rendered using *CurrUser* structure and then statically inserted into executed template.|Yes|
|`/login`|**loginGetHandler**|Used only to respond to GET request to render the login website.|No|
|`/login`|**loginPostHandler**|Used to handle received login data from user. If successful - redirects user to home page, otherwise displays error.|No|
|`/logout`|**logoutGetHandler**|Deletes current session and redirects user to */login* page.|No|
|`/register`|**registerGetHandler**|Used to render register website.|No|
|`/register`|**registerPostHandler**|Used to send data to backend and Redis to register user. Redirects to */login*.|No|
|`/docs`|**tutorialHandler**|Used to execute docs template.|No|
|`/positions`|**positionsGetHandler**|Used to render table with position data.|Yes|

## API routes
As of now, all API routes require auth.
|Route|Handler|Description|
|---|---|---|
|`/api/data`|**barDataHandler** | API endpoint used to retrieve data for bar charts displayed on home page.|
|`/api/positions`|**positionsHandler** |API endpoint which handles retriving positions data for each user at */positions* page.|
|`/api/update`|**updateEditHandler** | API endpoint handling PUT requests, which handle editing positions.|
|`/api/update`|**updateAddHandler** | API endpoint handling POST requests, which handle adding positions.|
|`/api/update`|**updateDeleteHandler** | API endpoint handling DELETE requests, which handle deleting positions.|
|`/api/month`|**monthSummaryHandler** | API endpoint handling requests from home page, retrieves month data from MongoDB.|

## Data structures
|Struct|Description|
|---|---|
|`WebUser`| Struct used to render username, dividend YTD and dividend tax on home page.|
|`MongoSummary`| Struct used to retrieve *YEAR_SUMMARY* from MongoDB.|
|`MonthData`| Simple struct used to handle one month data in form: **monthName - value** from Mongodb.|
|`MongoMonths`| Struct used to aggregate all months from MongoDB, uses *MonthData* structs to do that.|
|`DeletePosition`| Struct with one field, used to store ticker of position to be deleted.|