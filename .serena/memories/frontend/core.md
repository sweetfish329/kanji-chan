# Frontend Core

- Directory: `frontend/`
- Routing structure (SvelteKit routes):
  - `/` (root): Initial page containing event creation. Supports AI parsing input box.
  - `/admin`: Organizer panel to list events, adjust API keys, and trigger AI schedule suggest.
  - `/event/[uuid]`: Non-auth page for invitees to fill out their schedule coordinates.
- Styling resides in `src/app.css` using standard CSS variables and Glassmorphism aesthetics.