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

- create a top-level PipelineService (or IngestionService) that calls each step service in sequence. (the steps will be: 1. get the html [implemented in HtmlService], 2. process the HTML with the help of OpenAI [chapter "3. Component of processing the Data using OpenAI API."] 3. retrieve the zip files [not implemented yet] 4. open the zips into the buffer [not implemented yet] 5. process the buffer with the help of OpenAI into JSON [not implemented yet] 6. put the JSON into the DB [not implemented yet]) **Status:** IN_PROGRESS
- keep each step as its own service (e.g., HtmlService, OpenAIExtractService, ZipService, CsvParseService, DataService) for testability. **Status:** IN_PROGRESS

## Components 

### 0. Go structure with load conf. **Status:** IN_PROGRESS

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

### 2. Component of retrieving first HTML. **Status:** IN_PROGRESS

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

### 3. Component of processing the Data using OpenAI API.  **Status:** IN_PROGRESS

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
2. The link must start with "https://" and end with ".zip"
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


### 4. Component of retrieving retrieving the Zip and processing the data. **Status:** WAITING

### 5. Component of storing the data. **Status:** WAITING

### 6. Component of GET /data controller. **Status:** WAITING

# Technical details and Architecture
- High-level overview: Project is built up out of components. We build one by one, test and verify then LOCK
- Human (me) locks the components via specifing list in the spec
- Components: work as project plan says, only the IN_PROGRESS components

## OpenAI Usage Policy

OpenAI is used ONLY for:
- Extracting a ZIP download URL from raw HTML
- Parsing CSV rows into a predefined schema

Constraints:
- Deterministic output required
- Schema validation required on response
- Failure logged + retry max 3 times

## Interfaces

### External dependencies
- OpenAI API: 

## Data
- Postgres SQL

### Data model
- Db tables to create and use:
1. sources: ID as GUID, URL as text, comment as text
2. origin types: ID as GUID, origin as text (Onshore Wind/Hydropower/Solar/Thermal)
3. data: ID as GUID, region as text, country as text, auction volume as decimal.Decimal, sold volume as decimal.Decimal, weighted price as decimal.Decimal  
4. logs: ID as GUID, datetime, action as text (like DATA_RETRIVAL, ZIP_DOWNLOAD, OPENAPI_CALL_HTML_EXTRACT, OPENAPI_CALL_CSV_PARSE), outcome as text (SUCCESS, FAIL), message as text (if error, store the error data)

### Storage and migrations
- tables must be created if not exists, use internal/repo for function that migrates
