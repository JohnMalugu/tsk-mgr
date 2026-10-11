package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err := createSchema(db); err != nil {
		return nil, err
	}

	log.Println("Database connection established and schema initialized")
	return db, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	projectQuery := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		color TEXT NOT NULL DEFAULT '#6366f1',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(projectQuery)
	if err != nil {
		return err
	}

	taskQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		project_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		due_date DATETIME NOT NULL,
		estimate_minutes INTEGER,
		completed BOOLEAN NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		priority TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
	);`

	_, err = db.Exec(taskQuery)
	if err != nil {
		return err
	}

	// Safely attempt migration for existing databases
	_, _ = db.Exec(`ALTER TABLE tasks ADD COLUMN project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;`)

	return nil
}
