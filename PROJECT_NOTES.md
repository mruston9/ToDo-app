# Project: "Our ToDo List"

A full-stack todo app built as a learning exercise. This file is a snapshot of where
the project stands, meant as starting context for working on it (e.g. handing to Claude Code).

## Stack and hosting
- **Backend:** Node.js + Express (`server.js`)
- **Frontend:** single `index.html` file (vanilla HTML/CSS/JS, no framework)
- **Database:** PostgreSQL
- **Hosting:** Render (web service + Render-hosted Postgres). Live at https://todo-app-rxai.onrender.com
- **Source:** GitHub; deploys automatically on `git push origin main`

## Critical workflow gotchas (these have caused silent bugs before)
1. **Two databases** — a local Postgres (`todo_app`) and the Render one. Any schema change
   (`ALTER TABLE`, etc.) must be run **by hand on both** via `psql`. Forgetting the Render
   one causes silent failures (the app runs but the feature quietly doesn't work).
2. Code changes only go live after `git push` **and** Render finishes redeploying.
   Always confirm the deploy shows "Live" before testing.
3. `.env` holds the local `DATABASE_URL` and must never be committed. On Render,
   `DATABASE_URL` is set as an environment variable in the service settings.
4. Working habit: build one feature, test it end-to-end, then move on. Avoid batching
   several unrelated multi-layer changes at once.

## Database schema (`todos` table)
| Column        | Type         | Notes                                             |
|---------------|--------------|---------------------------------------------------|
| id            | serial       | primary key                                       |
| title         | varchar(255) | not null                                          |
| completed     | boolean      | default false                                     |
| created_at    | timestamp    | default now()                                     |
| position      | integer      | manual reordering                                 |
| completed_at  | timestamp    | set when checked complete, cleared when unchecked |
| assignee      | varchar(1)   | '' = unassigned, 'M' = Mark, 'D' = Dawn           |

## Backend endpoints (`server.js`)
- `GET /todos` — all todos, ordered by position
- `POST /todos` — add (placed at bottom via max position + 1)
- `PUT /todos/:id` — toggle complete; also sets/clears `completed_at`
- `DELETE /todos/:id` — hard delete (permanent, no soft-delete)
- `POST /todos/:id/move` — reorder up/down by swapping position with the neighbor
- `POST /todos/:id/assign` — cycle assignee '' -> M -> D -> ''

## Frontend features (`index.html`)
- Blue-grey theme, mobile-optimized for iPhone
- Each row: assignee badge (grey dash / blue M / orange D, tap to cycle) on the left,
  checkbox + title + timestamps in the middle, stacked up/down reorder arrows + red x
  delete on the right
- Titles are HTML-escaped on display (XSS protection) via an `escapeHtml` helper
- Delete asks for confirmation before the permanent delete
- "Hide completed / Show completed" toggle (frontend filter, resets on reload)
- Created timestamp on every item; green "Completed:" timestamp appears when done

## Known shortcuts / future work (don't fix unless asked)
- **No authentication** — the app is public; anyone with the URL can edit. Biggest known gap.
- Wide-open CORS, no rate limiting, no backend input validation.
- Delete is a hard delete with no recovery path.
- Inline `onclick` handlers with escaped titles — works but fragile; cleaner pattern is
  attaching event listeners in JS.
- Schema changes are hand-run on two databases (no migrations tool).
- Eventually want: filter by assignee (Mark/Dawn), possibly soft-delete.
