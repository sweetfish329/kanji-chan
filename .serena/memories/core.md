# Core

- `kanji-chan` is an AI-powered schedule coordination web application inspired by Chouseisan.
- It consists of a Go backend and a Svelte 5 frontend.
- Database state is managed in PostgreSQL via GORM.
- Main feature graph and entry points:
  - Backend code structure: `mem:backend/core`
  - Frontend code structure: `mem:frontend/core`
  - Global technology stack details: `mem:tech_stack`
  - Developer commands and system utilities: `mem:suggested_commands`
  - Code conventions and design patterns: `mem:conventions`
  - Task verification and pre-commit checks: `mem:task_completion`

## Project Invariants
- Event responders do NOT need to log in to submit candidate answers.
- Event organizers (organizers/kanjis) must authenticate via OAuth (Google or GitHub).
- Organizer-specific configurations, including their custom Gemini API keys, are stored in the database.