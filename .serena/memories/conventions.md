# Conventions

## Backend (Go)
- Project follows standard Go directory structure: `cmd/` for binaries, `internal/` for internal packages.
- Package breakdown:
  - `internal/model`: Database entity definitions.
  - `internal/database`: DB initialization, migration, and connection.
  - `internal/auth`: OAuth handler and session state via Goth.
  - `internal/ai`: Prompt orchestration and structured Gemini API calls.
  - `internal/handler`: REST endpoints using Echo.
- Error handling: return errors up to the handler layer to be converted to consistent JSON responses.

## Frontend (Svelte 5)
- Enforce Svelte 5 Runes: use `$state()`, `$derived()`, and `$props()` exclusively. Do NOT use legacy Svelte 4 store patterns or reactive assignments (`let` / `$:`).
- Styling: use Vanilla CSS. Avoid TailwindCSS unless explicitly requested.
- Network client: prefix API requests with `/api/` pointing to the Go server (proxied via Vite or set via `VITE_API_BASE_URL`).

## Authentication & Security
- Protect organizer routes using cookie-based sessions.
- Encrypt and store organizer's Gemini API key in the database, falling back to environment variables.