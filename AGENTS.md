# Repository Guidelines

It's a Golang backend.

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