# Task Completion

Before marking any task as complete, verify that the following checks pass:

## Backend Checks
- Format check: `go fmt ./...`
- Vet check: `go vet ./...`
- Tests: `go test ./...` runs and passes successfully.

## Frontend Checks
- Formatting: `bun run format` formats all Svelte and TypeScript files.
- Linting: `bun run lint` passes with no errors.
- TypeScript/Svelte Integrity: `bun run check` passes with no diagnostics errors.
- Build: `bun run build` completes successfully without errors.