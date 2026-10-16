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

	tagQuery := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		color TEXT NOT NULL DEFAULT '#10b981',
		created_at DATETIME NOT NULL,
		UNIQUE(user_id, name),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(tagQuery)
	if err != nil {
		return err
	}

	taskTagQuery := `
	CREATE TABLE IF NOT EXISTS task_tags (
		task_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		PRIMARY KEY (task_id, tag_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(taskTagQuery)
	if err != nil {
		return err
	}

	subtaskQuery := `
	CREATE TABLE IF NOT EXISTS subtasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		position INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(subtaskQuery)
	if err != nil {
		return err
	}

	commentQuery := `
	CREATE TABLE IF NOT EXISTS task_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(commentQuery)
	if err != nil {
		return err
	}

	timeEntriesQuery := `
	CREATE TABLE IF NOT EXISTS time_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		duration_seconds INTEGER NOT NULL DEFAULT 0,
		note TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(timeEntriesQuery)
	if err != nil {
		return err
	}

	taskDependenciesQuery := `
	CREATE TABLE IF NOT EXISTS task_dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		depends_on_task_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		UNIQUE(task_id, depends_on_task_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(taskDependenciesQuery)
	if err != nil {
		return err
	}

	activitiesQuery := `
	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		task_id INTEGER,
		action TEXT NOT NULL,
		summary TEXT NOT NULL,
		details_json TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(activitiesQuery)
	if err != nil {
		return err
	}

	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_time_entries_user_time ON time_entries(user_id, started_at);
	CREATE INDEX IF NOT EXISTS idx_time_entries_task_id ON time_entries(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_dependencies_task_id ON task_dependencies(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);
	CREATE INDEX IF NOT EXISTS idx_activities_user_created ON activities(user_id, created_at DESC);
	`

	_, err = db.Exec(indexQuery)
	return err
}
