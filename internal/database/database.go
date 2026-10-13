package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func InitDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库：%w", err)
	}

	if err := createTable(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title VARCHAR(200) NOT NULL,
		description TEXT,
		status VARCHAR(20) DEFAULT 'pending' CHECK(status IN ('pending', 'completed', 'cancelled')),
		priority VARCHAR(10) CHECK(priority IN ('low', 'medium', 'high')),
		due_date DATE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("创建任务表：%w", err)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)",
		"CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority)",
		"CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at)",
	}

	for _, index := range indexes {
		_, err := db.Exec(index)
		if err != nil {
			return fmt.Errorf("创建索引：%w", err)
		}
	}

	return nil
}
