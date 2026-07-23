package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Todo mirrors a row in the todos table.
// The `json:"..."` tags control the field names in the JSON sent to the browser.
// These MUST match what index.html expects, or the frontend silently breaks.
type Todo struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	Position    *int       `json:"position"`
	CompletedAt *time.Time `json:"completed_at"`
	Assignee    string     `json:"assignee"`
}

var pool *pgxpool.Pool

func main() {
	// Load .env if present. Ignore the error: on Render there's no .env file,
	// the variables come from the platform instead.
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	pool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// Route patterns include the HTTP method (Go 1.22+).
	// {id} is a wildcard, read back with r.PathValue("id").
	mux.HandleFunc("GET /todos", getTodos)
	mux.HandleFunc("POST /todos", addTodo)
	mux.HandleFunc("PUT /todos/{id}", toggleTodo)
	mux.HandleFunc("DELETE /todos/{id}", deleteTodo)
	mux.HandleFunc("POST /todos/{id}/move", moveTodo)
	mux.HandleFunc("POST /todos/{id}/assign", assignTodo)

	// Serve index.html and any other static files from the project folder.
	// Equivalent to app.use(express.static('.')).
	mux.Handle("/", http.FileServer(http.Dir(".")))

	log.Printf("Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// withCORS replaces the Express `cors` middleware.
// NOTE: "*" allows any site to call this API — same wide-open policy as the
// Node version. Tighten this when you add authentication.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- helpers -------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}

func idFrom(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

// --- handlers ------------------------------------------------------------

func getTodos(w http.ResponseWriter, r *http.Request) {
	rows, err := pool.Query(r.Context(),
		`SELECT id, title, completed, created_at, position, completed_at,
		        COALESCE(assignee, '')
		 FROM todos ORDER BY position`)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer rows.Close()

	todos := []Todo{} // not nil, so empty results encode as [] rather than null
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt,
			&t.Position, &t.CompletedAt, &t.Assignee); err != nil {
			writeErr(w, err)
			return
		}
		todos = append(todos, t)
	}
	writeJSON(w, http.StatusOK, todos)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}

	var t Todo
	err := pool.QueryRow(r.Context(),
		`INSERT INTO todos (title, completed, position)
		 VALUES ($1, false, COALESCE((SELECT MAX(position) FROM todos), 0) + 1)
		 RETURNING id, title, completed, created_at, position, completed_at,
		           COALESCE(assignee, '')`,
		body.Title).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt,
			&t.Position, &t.CompletedAt, &t.Assignee)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func toggleTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFrom(r)
	if err != nil {
		writeErr(w, err)
		return
	}

	var body struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}

	// Set the timestamp when checking, clear it when unchecking.
	var completedAt *time.Time
	if body.Completed {
		now := time.Now()
		completedAt = &now
	}

	var t Todo
	err = pool.QueryRow(r.Context(),
		`UPDATE todos SET completed = $1, completed_at = $2 WHERE id = $3
		 RETURNING id, title, completed, created_at, position, completed_at,
		           COALESCE(assignee, '')`,
		body.Completed, completedAt, id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt,
			&t.Position, &t.CompletedAt, &t.Assignee)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFrom(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	if _, err := pool.Exec(r.Context(), `DELETE FROM todos WHERE id = $1`, id); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func moveTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFrom(r)
	if err != nil {
		writeErr(w, err)
		return
	}

	var body struct {
		Direction string `json:"direction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, err)
		return
	}

	ctx := r.Context()

	// A transaction: both position swaps succeed together, or neither does.
	// The Node version did these as two separate statements, which could in
	// principle leave the order half-updated. This is a small improvement.
	tx, err := pool.Begin(ctx)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer tx.Rollback(ctx) // no-op if the commit below succeeds

	var currentPos int
	if err := tx.QueryRow(ctx,
		`SELECT position FROM todos WHERE id = $1`, id).Scan(&currentPos); err != nil {
		writeErr(w, err)
		return
	}

	neighbourQuery := `SELECT id, position FROM todos WHERE position > $1 ORDER BY position ASC LIMIT 1`
	if body.Direction == "up" {
		neighbourQuery = `SELECT id, position FROM todos WHERE position < $1 ORDER BY position DESC LIMIT 1`
	}

	var adjID, adjPos int
	if err := tx.QueryRow(ctx, neighbourQuery, currentPos).Scan(&adjID, &adjPos); err != nil {
		// No neighbour means it's already at the top or bottom — nothing to do.
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
		return
	}

	if _, err := tx.Exec(ctx, `UPDATE todos SET position = $1 WHERE id = $2`, adjPos, id); err != nil {
		writeErr(w, err)
		return
	}
	if _, err := tx.Exec(ctx, `UPDATE todos SET position = $1 WHERE id = $2`, currentPos, adjID); err != nil {
		writeErr(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func assignTodo(w http.ResponseWriter, r *http.Request) {
	id, err := idFrom(r)
	if err != nil {
		writeErr(w, err)
		return
	}

	var current string
	if err := pool.QueryRow(r.Context(),
		`SELECT COALESCE(assignee, '') FROM todos WHERE id = $1`, id).Scan(&current); err != nil {
		writeErr(w, err)
		return
	}

	// Cycle: '' -> M -> D -> ''
	next := ""
	switch current {
	case "":
		next = "M"
	case "M":
		next = "D"
	}

	var t Todo
	err = pool.QueryRow(r.Context(),
		`UPDATE todos SET assignee = $1 WHERE id = $2
		 RETURNING id, title, completed, created_at, position, completed_at,
		           COALESCE(assignee, '')`,
		next, id).
		Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt,
			&t.Position, &t.CompletedAt, &t.Assignee)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}
