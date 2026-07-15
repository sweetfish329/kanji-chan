# Suggested Commands

## Database
- Start PostgreSQL database: `docker compose up -d db` or `podman-compose up -d db`
- Start entire stack: `docker compose up -d`

## Backend (in /backend)
- Test: `go test ./...`
- Build: `go build -o server.exe ./cmd/server`
- Run: `go run ./cmd/server/main.go`

## Frontend (in /frontend)
- Install: `bun install`
- Dev Server: `bun run dev`
- Static Type Check: `bun run check`
- Lint: `bun run lint`
- Format: `bun run format`

## OS Invariants
- Development is carried out on Windows.
- Ensure path configurations in scripts support both backslashes (\) and forward slashes (/) appropriately.