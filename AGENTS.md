# Repository Guidelines

It's a Golang backend.

## Editing rules
- SPEC.md contains software development tasks and descriptions how to complete these tasks. By default work with the tasks that have **Status:** IN_PROGRESS. Amendments that have **Status:** IN_PROGRESS are also marked to work with in current iteration.
- Tasks that have **Status:** WAITING haven't formed yet. They might provide context but the quality of the information is very low or missing.
- Without consent, do not modify parts of the code that was created by tasks and descriptions marked as **Status:** DONE_AND_LOCKED. Asking consent is very okay. These tasks and descriptions are already implemented. Some of these have ammendments, so if you use them as a context, check for relevant amendments too. 
- When asking permissions, make a numbered list of required modifications, so I can replay like: 1. yes 2. yes 3. no 4. yes  

## Specs

Detailed instructions about project spec and tasks 
→ `docs/SPEC.md`

## Type Safety
- Do not invent new types if Go already has it
- No use of `interface{}`
- No unchecked type assertions
- No nil dereference tolerated

## Type & Nullability Rules (Global)

- Pointer types (`*T`) represent nullable columns
- Value types (`T`) represent NOT NULL columns
- `sql.Null*` types are forbidden unless explicitly approved
- JSON `omitempty` does NOT imply database nullability
- GORM auto-migration must not infer nullability implicitly
- When change is required in components that are DONE_AND_LOCKED, ask permission 

## Validation before Pull Request

- Run go test ./... after every change
- Use Go linter golangci-lint run
- go vet ./...

Regression checklist:
- [ ] No type assertions added
- [ ] Existing tests unchanged or strengthened
- [ ] Public function signatures unchanged
- [ ] No behavior change unless explicitly requested

## Code conventions

- _test.go files are next to function files
- main.go describes API routes, /cmd/ folder
- controllers are in /cmd/controllers/
- secrets.json is only for dev, in k8s its handled properly as a secret