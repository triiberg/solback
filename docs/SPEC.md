# SPEC

## Summary
- Problem: European Energy Exchange have auction category "Guarantees of Origin" but the data is stored as CSV files that are packed into zip and are very hard to consume. Can't build frontend straight on top of it.
- Proposed solution: Read data from website specified in db table sources (url=https://www.eex.com/en/markets/energy-certificates/french-auctions-power) and extract a link of zip file from the HTML. That zip contains data of energy certificate prices. Data must be extracted and stored into the database. The controller of this backend will provide REST API of the data. 
- Success criteria: The backend retrieves data, stores and servers HTTP calls for frontend (other project)

## Goals
- Cron (or API call) triggers process to get the HTML
- The data from HTML is processed by OpenAI API to find link with text  "GO YYYY-YYYY Global Results" like "GO 2024-2025 Global Results. The href is to ZIP file. Return the link of href.
- The zip file of the link must be downloaded and unpacked using buffer.
- Contents of the archive are CSV files with actual data and must read into the buffer. 
- The data will be passed to the OpenAI API with schama that specifies the data format. 

## Non-goals
- Trying to filter out the data locally
- Provide frontend
- Store Zip files or CSV files
- Swagger not needed, only 4 endpoints returning JSON: GET /health, GET /logs, GET /sources and GET /data?...

# Project Specification

## Status Legend
- WAITING - requirements incomplete, ambiguous, or unvalidated
- IN_PROGRESS - actively being implemented
- DONE_AND_LOCKED - implemented, tested, and must not be modified

## Components 

### 0. Go structure with load conf. **Status:** IN_PROGRESS

####  1. Dependencies:
- Cron (create on example task that runs once in hour, prints out "hello")
- Gin for controllers (cmd/main.go , cmd/controllers/health.go)
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


### 1. Conteinerization. **Status:** WAITING

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




### 2. Component of retrieving first HTML. **Status:** WAITING

### 3. Component of processing the Data using OpenAI API.  **Status:** WAITING

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


## Observability
- Logging:
- Metrics:
- Tracing:
- Alerts:

## Error handling
- 

## Performance and scalability
- 

## Testing
- Unit:
- Integration:
- E2E:

## Rollout
- Deployment plan:
- Feature flags:
- Backward compatibility:

## Risks
- logs can reveal credentials

## Open questions
- 

## Decisions
- 

