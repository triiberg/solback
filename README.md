# Instructions for humans

## Docker scrapbook

### Build Docker

docker compose up --build -d

### If you want a clean restart:

docker compose down
docker compose up --build -d

### If you only changed the app image (not DB), you can do:

docker compose up --build -d api

## Check test coverage

go test ./... -cover

go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# Run codex

## Initial prompt

Generate a file named AGENTS.md that serves as a contributor guide for this repository.
Your goal is to produce a clear, concise, and well-structured document with descriptive headings and actionable explanations for each section.
Follow the outline below, but adapt as needed — add sections if relevant, and omit those that do not apply to this project.

Document Requirements

- Title the document "Repository Guidelines".
- Use Markdown headings (#, ##, etc.) for structure.
- Keep the document concise. 200-400 words is optimal.
- Keep explanations short, direct, and specific to this repository.
- Provide examples where helpful (commands, directory paths, naming patterns).
- Maintain a professional, instructional tone.

Recommended Sections

Project Structure & Module Organization

- Outline the project structure, including where the source code, tests, and assets are located.

Build, Test, and Development Commands

- List key commands for building, testing, and running locally (e.g., npm test, make build).
- Briefly explain what each command does.

Coding Style & Naming Conventions

- Specify indentation rules, language-specific style preferences, and naming patterns.
- Include any formatting or linting tools used.

Testing Guidelines

- Identify testing frameworks and coverage requirements.
- State test naming conventions and how to run tests.

Commit & Pull Request Guidelines

- Summarize commit message conventions found in the project’s Git history.
- Outline pull request requirements (descriptions, linked issues, screenshots, etc.).

(Optional) Add other sections if relevant, such as Security & Configuration Tips, Architecture Overview, or Agent-Specific Instructions.

## Prompt with mendments

docs/SPEC.md "Amendments" sectioction describes need for config.json and internal/repo/repo.go must add data if sources table is empty. Docker container must have that file too. Everything else works. Follow AGENTS.md for best practices and add only missing part described in docs/SPEC.md "Amendments" section.

## Prompt 4

follow AGENTS.md to work with project's SPEC.md. We're currently working on compenent "### 2. Component of retrieving first HTML. **Status:** IN_PROGRESS". This part of the project must create a methods to write and read logs. It will be possible to observe the work of the program using the endpoint GET /logs. The true work that the program must do is to retrieve the data from URL(s) stored in the database. Observe the guidelines, guidrails and follow similar code convention as the existing code.

Previous components are not in the scope (the backend runs well already, the sources table exists and getSources exists in SourceService). The next step of processing the retrieved HTMLs are not in the scope of this iteration as well. 

If need to rewrite anything out of the scope, please notify explicitly.
