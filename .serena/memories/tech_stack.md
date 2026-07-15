# Tech Stack

## Backend
- Go 1.23+ (configured as 1.26 in `go.mod`)
- Echo v4 (Web Framework)
- GORM (PostgreSQL ORM)
- Goth (OAuth 2.0 multi-provider library)
- Google GenAI Go SDK (v1.63.0) for Gemini API

## Frontend
- Svelte 5 (utilizing latest Runes)
- SvelteKit 2 (routing & app framework)
- Vite 8 (bundler & dev server)
- Bun (package manager and runtime environment)
- oxlint (Linter) & oxfmt (Formatter)
- @zerodevx/svelte-toast (notification UI)
- dayjs (date parsing and formatting)

## Database & Infrastructure
- PostgreSQL 16 (running in Docker Compose / podman-compose)