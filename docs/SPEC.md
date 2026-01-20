# SPEC

## Summary
- Problem: European Energy Exchange have auction category "Guarantees of Origin" but the data is stored as CSV files that are packed into zip and are very hard to consume. Can't build frontend straight on top of it.
- Proposed solution: Read data from website specified in db table sources (url=https://www.eex.com/en/markets/energy-certificates/french-auctions-power) and extract a link of zip file from the HTML. That zip contains data of energy certificate prices. Data must be extracted and stored into the database. The controller of this backend will provide REST API of the data. 
- Success criteria: The backend retrieves data, stores and servers HTTP calls for frontend (other project)

## Goals
- Cron (or API call) triggers the pipeline service that calls out other services:
1. SourceService has functions that read the sources
2. HtmlService downloads content from these sources
3. Every HTML that was received must be preprocessed, injected into the prompt and sent to the OpenAI API by the OpenAiService, then result of links wil be given next step
4. The ZIP file(s) must be downloaded into buffer and extracted by ZipService
5. Again OpenAiService comes into action and second prompt with the data will convert CSV into valid JSON
6. At the end of the pipeline DataService will come and store it into the database
- There are endpoint that helps to observe, trigger refresh and ultimately display the complete dataset. 

## Non-goals
- Trying to filter out the data locally
- Provide frontend
- Store Zip files or CSV files
- Swagger not needed

# Project Specification

## Status Legend
- WAITING - requirements incomplete, ambiguous, or unvalidated
- IN_PROGRESS - actively being implemented
- DONE_AND_LOCKED - implemented, tested, and must not be modified

## Amendments
- add config.json (store default source data URL: "https://www.eex.com/en/markets/energy-certificates/french-auctions-power" and comment "Default source") **Status:** DONE_AND_LOCKED
- make sure config.json is in .gitignore **Status:** DONE_AND_LOCKED
- make sure config.json is available in Docker container **Status:** DONE_AND_LOCKED
- repo.go must check if sources table is empty add url and comment from config.json **Status:** DONE_AND_LOCKED

- create a top-level PipelineService (or IngestionService) that calls each step service in sequence. (the steps will be: 1. get the html [implemented in HtmlService], 2. process the HTML with the help of OpenAI [chapter "3. Component of processing the Data using OpenAI API."] 3. retrieve the zip files [not implemented yet] 4. open the zips into the buffer [not implemented yet] 5. process the buffer with the help of OpenAI into JSON [not implemented yet] 6. put the JSON into the DB [not implemented yet]) **Status:** DONE_AND_LOCKED
- keep each step as its own service (e.g., HtmlService, OpenAIExtractService, ZipService, CsvParseService, DataService) for testability. **Status:** DONE_AND_LOCKED

- logs model must have event identifier that is given by PipelineService when starting the process. This will allow me to filter out log events per session. **Status:** DONE_AND_LOCKED
- PipelineService when triggered by cron or refresh endpoint, generates new eventId and assignes it to log records that are created in this session of refreshing the data. **Status:** DONE_AND_LOCKED
- logs endpoint must allow filter eventId. **Status:** DONE_AND_LOCKED
- when creating log record of OPENAPI_CALL_CVS_PARSE, attache the filename that is currently parsed. **Status:** DONE_AND_LOCKED
- log service must have truncate logs function. **Status:** DONE_AND_LOCKED
- log controller must have DELETE /logs endpoint that calls truncate logs function from logs service. **Status:** DONE_AND_LOCKED
- refresh endpoint must be asynchronus: just run it as go routine. **Status:** DONE_AND_LOCKED

- all data aquired when OPENAPI_CALL_CSV_PARSE is SUCCESS must be in the JSON that will be parsed into model that will be stored in the database. Currently if any of the OpenAPI calls that parse CSV fails, nothing is available in database data table - but must be. **Status:** DONE_AND_LOCKED

- no reason to pass the same CSV to OpenAI if it failed: the input is the same and fail next time. One attempt only.   **Status:** DONE_AND_LOCKED

- add function to the CSV service, that extracts year and month from filename.  **Status:** DONE_AND_LOCKED
- change the schema sent to OpenAI when parsing the CSV: do not require year and month properties to be added by OpenAI. Add year and month when the payload comes back.  **Status:** DONE_AND_LOCKED


## Components 

### 0. Go structure with load conf. **Status:** DONE_AND_LOCKED

####  1. Dependencies:
- Cron (create on example task that runs once in hour, prints out "hello")
- Gin for controllers (cmd/main.go, cmd/controllers/health.go, etc)
- Gorm for migration (internal/repos/repo.go - func Connect() and func Migrate())

#### 2. Folder structure:
- cmd
- secrets.json (store db dsn and OpenAPI key) 
- internal/models (1 model in this step: sources)
- internal/repo (migrate model "sources" and "logs")
- internal/services (sourceService, func getSources)
- cmd/controllers (healt and sources endpoints)

#### 3. Expected result:
- file structure 
- I can run go mod tidy to download required dependencies
- secrets.json with two records
- main.go starts up
- GET /health endpoint responds with code 200
- cron prints "hello" to the console every hour
- test coverage for controllers, _test.go file next to actual functionality
- main.go excluded from coverage


### 1. Conteinerization. **Status:** DONE_AND_LOCKED

#### 1. Containers

##### Go API
Artifact:
- `Dockerfile`

Responsibilities:
- Build Go backend binary
- Run HTTP API

Must NOT:
- Run a database
- Include DB binaries

---

##### Database
Artifact:
- Docker image (official)

Responsibilities:
- Persist application data

---

#### 2. Local Orchestration
Artifact:
- `docker-compose.yml`

Requirements:
- Start API and DB as separate containers
- API connects to DB via service name

#### 3. Secrets:
- add dsn into secrets.json

#### 4. Expected result:
- can run `go run cmd/main.go`

### 2. Component of retrieving first HTML. **Status:** DONE_AND_LOCKED

#### 1. Objectives
- get URLs from sources table
- loop through the source url entries and make HTTP GET request against every URL
- read the contents into string (output to console)
- the HTML retrieval mechanism must be triggered by Cron every hour or via controller GET /refresh
- every round of data retrieval must be logged into log table: did the request succeeded or failed with datetime in UTC.
- the fact of triggered data retrieval is observable through GET /logs endpoint

#### 2. Coding
- create LogService similar to SourceService that has functions to filter out n last records (order by datetime latest records first, limit by input parameter)
- create GET /logs endpoint (controller) that is using LogService to get last n records (query param, default 20), JSON array of JSON objects
- create service HtmlService, similar structure as SourceService
- create functions that read slice of the URLs and make HTTP Get request to the URLs
- the looper function must log into log table when it started
- the looper function must log the HTTP GET call result code, was it successful or not with datetime in UTC 

#### 3. Expected results
- cron triggers process that retrieves data from URL(s) specified in sources table
- GET /refresh endpoint tiggers process that retrieves data from URL(s) specified in source table, same process as the cron would trigger
- the log entries were written into log table: 1. the fact that loop started 2. the result of http request
- the log entries are observable using Get /logs end point with parameter `?n=<number of records>`
- by default GET /log endpoint returns 20 rows in JSON object
- test coverage for creating logs
- test coverate to retrieve logs
- test coverage to loop though the sources and retrievening the HTML from these URLs
- all tests passed
- its bossible to rebuild Docker image and run it

### 3. Component of processing the Data using OpenAI API.  **Status:** DONE_AND_LOCKED

#### 1. Objectives
- the data from HtmlService download will be passed to OpenAiService (the )
- first the prefiltering phase must be applied that strips irrelavant tags
- the second phase is to pass the HTML junk to OpenAI with instructions to parse the messi HTML and find links to ZIP files (return JSON object of links) or if no relevant ZIP files detected, returns an error (JSON)
- the process of getting result or error must be written into logs table and be observable through the GET /logs endpoint (endpoint works great, no need to modify logs controller)

#### 2. Coding
1. the OpenAi integration must use the API Key (openai_api_key) from secrets.json
2. Create OpenAiService, similar to other services in this project
3. Create prefiltering function that will take in large HTML file (example: docs/example.html) and returns table elements that holds links to various ZIP files.
4. OpenAiService must have a function that takes in string of HTML (the example is ./docs/example.html) and instructions to find the link of most recent ZIP of "GO year1-year2 Global Results" like "GO 2024-2025 Results" and return JSON. Instruction in following paragraph "#### 4. Instructions for OpenAI".
5. The results of OpenAI API must be JSONs. It could be result or an error. Both outcomes must be logged.
6. Create struct for following JSON { "error": "string", period: "string" description: "string", "link": "string" }
7. Use JSON structure when retrieving the results like { "error": "", period: "20..-20.." description: "GO .... results", "link": "https:// .... .zip" } or { "error": "EMPTY_HTML" }. If error "" it was successful result. 
8. Log the result into logs table

#### 3. Instructions for OpenAI
Prompt to OpenAI must be 
```
Non-negotiable rules:
1. Return only valid JSON
2. Use JSON structures given: result
3. If no result, return { "error": "NO_RESULTS", "period": "", "description": "", "link": "" } or { "error": "EMPTY_HTML", ... }
4. If solid match found return { "error": "", period: "20..-20.." description: "GO .... results", "link": "https:// .... .zip" }
5. If found more than one result, return the one with greatest year number wins.

Instructions:
Find link to the most relevant ZIP file. Known criterias 
1. The description must say GO or Guarantee of Origin, the year number(s) and states that these are the "results"
2. The link must end with ".zip"
3. Return result in form described in rules section

Notes:
1. The following HTML is already stripped and might not be a valid HTML
2. It is prechecked it does have links and string "zip" is appears in the content but no quarantee its a link

HTML:
<REPLACE_WITH_HTML_STRING>
```

#### 4. Expected results
1. At first table elements of the raw HTML is extracted 
2. the HTML of the table elents is injected into the prompt
3. prompt is called using OpenAI API with specified key found in the secrets.json
4. when OpenAI returns results log the outcome into logs

#### 5. Extra requirements
- amendments of creating a pipeline service that orchestrates other service calls are implemented
- double check if main.go with cron and refresh endpoints still work: the HtmlService no accessed by cron or controller but from PipelineService
- high test coverage of all code created
- test that can parse the ./docs/example.html and return the table elements 
- test that can be run with help of the table extraction function
- the app is runnable/compilable


### 4. Component of retrieving retrieving the Zip and processing the data. **Status:** DONE_AND_LOCKED

#### 1. Objectives
- this step must be called by PipelineService after the link is retrieved by previous step
- pipeline service invokes functions from ZipService with data from pervious step (JSON of a link to a Zip file and source URL)
- Zip file download

#### 2. Coding
- create ZipService similar to other services
- create a function that takes in the link and source and validates the link
- if needed using the source full URL must be created (http......zip)
- creates HTTP call and retrieves the bytes of the ZIP file
- logs success / error
- passes the buffer to the next step (described as 5. Processing the Zip file locally and with help of OpenAI API.)

#### 3. Expected results
1. Zip file link is validated
2. Zip is downloaded into the buffer
3. The event of download is observable in the logs table
4. Tests coverage over the code has been created in this iteration

### 5. Processing the Zip file locally and with the help of OpenAI API. **Status:** DONE_AND_LOCKED

#### 1. Objectives
- this current zip processing step must be called by PipelineService after the ZIP is successfully downloaded and loaded into the buffer 
- the byte buffer of the zip file will be processed into the junks possible to send to OpenAI API
- precheck: validate OpenAI model token budget and max rows per request using synthetic XLSX-like payload (no real data, no ingestion) - estimate if its possible to get results with planned data, write into logs
- processed data is passed to the OpenAI api
- returned JSON must be validated
- all buffers and memory that were temporary allocated to work with bytes of the ZIP and excel files are freed
- the JSON that OpenAI returned will be passed to the next step (described in 6. Component of storing the data.)

#### 2. Example data
1. ./docs/ folder has example zip file 20251119_GO_2024_2025_GLOBAL_Results.zip 
2. ./docs/ folder has example xlsx files that represent the data in the zip file: 20250520_February_2025_77_GLOBAL_Results_detailedresults.xlsx and 20251119_August_2025_83_GLOBAL_Results_detailedresults.xlsx

#### 3. Coding
1. Unpack ZIP from memory
   - Use `archive/zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))`.
   - Do not write archive contents to disk.
2. Select XLSX entries
   - Iterate `zipReader.File`.
   - Skip directories, `__MACOSX`, and non-`.xlsx` files.
3. Read + parse XLSX
   - Read each XLSX entry into memory.
   - Open workbook from an `io.Reader` (e.g., excelize `OpenReader`).
   - Select target sheet (default: first sheet; fallback: scan sheets to find a title cell containing "Aggregated Auction Results").
   - Extract rows as `[][]string`.
4. Deterministically extract structure
   - Identify metadata rows (e.g., "Number of Participants to the Auction") and parse the integer value.
   - Identify the main table header row by detecting presence of both "Region" and "Technology" (FR/EN variants).
   - Data rows start after the header row; stop at the first block of empty/invalid rows.
5. Prepare OpenAI input payload (no raw XLSX)
   - Build `{source_file, participants, headers, rows}` from extracted sheet rows.
6. Call OpenAI with Structured Outputs (strict JSON schema)
   - Enforce schema compliance; reject invalid output.

#### 4. OpenAI structured extraction contract

- The OpenAI request MUST include a strict JSON Schema (Structured Outputs).
- Free-form JSON responses are not allowed.
- `additionalProperties` is set to `false` to prevent hallucinated fields.
- The model is responsible only for:
  - Mapping headers to canonical field names
  - Converting decimal commas to decimal points
  - Converting "-" or empty cells to null
  - Coercing numeric values
- The backend MUST reject any response that:
  - Fails JSON decoding
  - Violates the schema   

Schema:
```
{
  "name": "auction_results",
  "strict": true,
  "schema": {
    "type": "object",
    "properties": {
      "source_file": {
        "type": "string",
        "description": "Original XLSX file name",
        "year": "part of the filename",
        "month": "part of the filename"
      },
      "participants": {
        "type": "integer",
        "description": "Number of participants in the auction"
      },
      "rows": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "year": {
              "type": "number"
            },
            "month": {
              "type": "number"
            },
            "region": {
              "type": "string"
            },
            "technology": {
              "type": "string"
            },
            "total_volume_auctioned": {
              "type": "number"
            },
            "total_volume_sold": {
              "type": "number"
            },
            "weighted_avg_price_eur_per_mwh": {
              "type": "number"
            }
          }
        }
      }
    },
    "required": ["rows"],
    "additionalProperties": false
  }
}
```
#### 3. Expected results
1. Unpacked zip is extracted and passed to OpenAI
2. The result of OpenAI is retrieved 
3. The fact of success or failure is logged into logs table
4. Tests coverage over the code has been created in this iteration

### 6. Component of storing the data. **Status:** DONE_AND_LOCKED

#### 1. Objectives
- Data model to hold the data processed by OpenAI must be created
- Migration script for repo creates the DB table if does not exist
- New DataService must be created with functions to store and retrieve data in the database
- OpenAI results must be parsed into the model
- Results must be stored in the database
- logged how many rows added

#### 2. Coding
- create model of data
- create migration into repo.go
- create DataService similar to other services
- create functions to be called by PipeLine service with the payload from previous steps
- log success or fail

#### 3. Expected results
- structures to deal with the data has been created
- PipelineService calls the storing functions and saves data to the database
- events are logged
- test were created about the data storing
- make sure the PipelineService could run through the flow from the point it starts with the source URLS until the end where the actual data is saved into the database.


### 7. Component of GET /data controller and DELETE /data controller. **Status:** DONE_AND_LOCKED

#### 1. Objectives
- endpoint to retrieve data
- andpoint to delete data

#### 2. Expected results
- can GET /data with filters: period and technology
- can DELETE /data (truncates everything)

### 8. Create table to avoid duplications: one file can be processed only once **Status:** IN_PROGRESS

#### 1. Objectives
- no ZIP file must be parsed twice.
- no duplicate auction results in the DB.

#### 2. Coding
- create model and db table to store filenames of ZIP files.
- create function to store ZIP filenames.
- create function into html service to check if this particular file is already processed.
- if file already downloaded and processed, log the fact and exit.

#### 3. Expected Results
- no duplicate values in the auction results table / output.
- no wasting time of processing ZIP files that are already downloaded and processed.
- keeping record of files that are already processed.
- old tests must be fixed
- new functionality must be covered with tests
- project can be started up and built into Docker container
- all tests were passed

# Technical details and Architecture
- High-level overview: Project is built up out of components. We build one by one, test and verify then LOCK
- Human (me) locks the components via specifing list in the spec
- Components: work as project plan says, only the IN_PROGRESS components

## OpenAI Usage Policy

OpenAI is used ONLY for:
- Extracting a ZIP download URL from raw HTML
- Parsing CSV rows into a predefined schema

## Interfaces

### External dependencies
- OpenAI API

## Data
- Postgres SQL

### Data model
- Db tables to create and use:
1. sources: ID as GUID, URL as text, comment as text
2. origin types: ID as GUID, origin as text (Onshore Wind/Hydropower/Solar/Thermal)
3. data: ID as GUID, region as text, country as text, auction volume as decimal.Decimal, sold volume as decimal.Decimal, weighted price as decimal.Decimal  
4. logs: ID as GUID, eventId as GUID, datetime, action as text (like DATA_RETRIVAL, ZIP_DOWNLOAD, OPENAPI_CALL_HTML_EXTRACT, OPENAPI_CALL_CSV_PARSE), outcome as text (SUCCESS, FAIL), message as text (if error, store the error data)

### Storage and migrations
- tables must be created if not exists, use internal/repo for function that migrates
