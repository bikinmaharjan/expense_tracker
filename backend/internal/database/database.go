package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) *sql.DB {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal(err)
	}

	// Create tables if they don't exist
	if err = createTables(db); err != nil {
		log.Fatal(err)
	}

	return db
}

func createTables(db *sql.DB) error {
	// Create tags table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			color TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create payments table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			info TEXT NOT NULL,
			amount REAL NOT NULL,
			date_paid DATE NOT NULL,
			fully_paid BOOLEAN DEFAULT false,
			invoice_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create payment_tags junction table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payment_tags (
			payment_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (payment_id, tag_id),
			FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create documents table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			file_path TEXT NOT NULL,
			original_name TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create document_tags junction table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS document_tags (
			document_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (document_id, tag_id),
			FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_payments_date ON payments(date_paid);
		CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
		CREATE INDEX IF NOT EXISTS idx_documents_title ON documents(title);
	`)
	if err != nil {
		return err
	}

	return nil
}
