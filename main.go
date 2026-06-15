package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskInput struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type App struct {
	db *sql.DB
}

func main() {
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &App{db: db}

	if err := app.createTable(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r * http.Request) {
    http.ServeFile(w, r, "index.html")
})
	mux.HandleFunc("GET /tasks", app.listTasks)
	mux.HandleFunc("POST /tasks", app.createTask)
	mux.HandleFunc("GET /tasks/{id}", app.getTask)
	mux.HandleFunc("PUT /tasks/{id}", app.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", app.deleteTask)

	// Middleware de logging
	loggedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		mux.ServeHTTP(w, r)
	})

	// Configuração da porta via variável de ambiente
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Servidor com timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggedMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Servidor rodando em http://localhost:%s", port)
	log.Println("Teste com: curl http://localhost:" + port + "/tasks")
	log.Fatal(srv.ListenAndServe())
}

func (app *App) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		done INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := app.db.Exec(query)
	return err
}

func (app *App) listTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query(`
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (app *App) createTask(w http.ResponseWriter, r *http.Request) {
	input, ok := readTaskInput(w, r)
	if !ok {
		return
	}

	result, err := app.db.Exec(`
		INSERT INTO tasks (title, done)
		VALUES (?, ?)
	`, input.Title, input.Done)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	task, err := app.findTask(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (app *App) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	task, err := app.findTask(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "tarefa nao encontrada")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (app *App) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	input, ok := readTaskInput(w, r)
	if !ok {
		return
	}

	result, err := app.db.Exec(`
		UPDATE tasks
		SET title = ?, done = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, input.Title, input.Done, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "tarefa nao encontrada")
		return
	}

	task, err := app.findTask(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (app *App) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	result, err := app.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if affected == 0 {
		writeError(w, http.StatusNotFound, "tarefa nao encontrada")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *App) findTask(id int64) (Task, error) {
	row := app.db.QueryRow(`
		SELECT id, title, done, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id)

	return scanTask(row)
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (Task, error) {
	var task Task
	var done int
	var createdAt string
	var updatedAt string

	err := scanner.Scan(&task.ID, &task.Title, &done, &createdAt, &updatedAt)
	if err != nil {
		return Task{}, err
	}

	task.Done = done == 1
	task.CreatedAt = parseSQLiteTime(createdAt)
	task.UpdatedAt = parseSQLiteTime(updatedAt)

	return task, nil
}

func parseSQLiteTime(value string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		t, err := time.Parse(format, value)
		if err == nil {
			return t
		}
	}

	return time.Time{}
}

func readID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "id invalido")
		return 0, false
	}

	return id, true
}

func readTaskInput(w http.ResponseWriter, r *http.Request) (TaskInput, bool) {
	var input TaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "json invalido")
		return TaskInput{}, false
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title e obrigatorio")
		return TaskInput{}, false
	}

	return input, true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}