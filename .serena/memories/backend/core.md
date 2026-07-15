# Backend Core

- Directory: `backend/`
- Implements Chouseisan clone database models using GORM:
  - `User`: Authenticated event organizer. Stores encrypted Gemini API key.
  - `Event`: The coordinate event (UUID primary key).
  - `EventCandidate`: Candidate slot for the event.
  - `Response`: Response header containing the responder name and comment.
  - `CandidateAnswer`: Answers ('ok', 'maybe', 'ng') mapped to each candidate slot.
- AI Integration:
  - Parse Text: `POST /api/ai/parse-event` extracting title, description, and timeslots into structured JSON.
  - Suggest Schedule: `POST /api/ai/suggest-schedule` providing ranking and justifications based on answers and preferences.