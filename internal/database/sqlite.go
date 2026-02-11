package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/1it/go-submission-service/internal/models"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	log.Printf("Opening SQLite database at %s", dbPath)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("ERROR: Failed to open SQLite database: %v", err)
		return nil, err
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)                 // Maximum number of open connections
	db.SetMaxIdleConns(5)                  // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Printf("ERROR: Failed to ping database: %v", err)
		return nil, err
	}

	log.Println("SQLite database connection established successfully")
	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Close() error {
	log.Println("Closing database connection")
	return s.db.Close()
}

func (s *SQLiteDB) Initialize() error {
	log.Println("Initializing database schema using migrations")

	// Run all pending migrations
	if err := s.Migrate(); err != nil {
		log.Printf("ERROR: Failed to run migrations: %v", err)
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// AddSubscriber adds a new email or updates the timestamp if it exists.
func (s *SQLiteDB) AddSubscriber(email string) error {
	log.Printf("Adding or updating subscriber: %s", email)
	query := `
    INSERT INTO subscribers (email) 
    VALUES (?) 
    ON CONFLICT(email) DO UPDATE SET 
        updated_at = CURRENT_TIMESTAMP
    `

	result, err := s.db.Exec(query, email)
	if err != nil {
		log.Printf("ERROR: Failed to add/update subscriber %s: %v", email, err)
		return err
	}

	rows, _ := result.RowsAffected()
	log.Printf("Subscriber %s operation successful: %d rows affected", email, rows)
	return nil
}

// GetSubscriber retrieves a subscriber by email. Returns sql.ErrNoRows if not found.
func (s *SQLiteDB) GetSubscriber(email string) (*models.Subscriber, error) {
	log.Printf("Looking up subscriber by email: %s", email)
	query := `SELECT id, email, status, created_at, updated_at FROM subscribers WHERE email = ?`

	row := s.db.QueryRow(query, email)

	var subscriber models.Subscriber
	// SQLite stores timestamps as strings, need to scan into string first
	var createdAtStr, updatedAtStr string

	err := row.Scan(&subscriber.ID, &subscriber.Email, &subscriber.Status, &createdAtStr, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Subscriber not found: %s", email)
		} else {
			log.Printf("ERROR: Failed to retrieve subscriber %s: %v", email, err)
		}
		return nil, err // Returns sql.ErrNoRows if not found
	}

	// Parse timestamp strings. Assuming UTC or format stored by CURRENT_TIMESTAMP
	// Using time.RFC3339Nano or a specific format might be necessary depending on exact storage.
	// For simplicity, let's try a common format. Consider error handling for Parse.
	createdAt, parseErr := time.Parse(time.RFC3339, createdAtStr)
	if parseErr != nil {
		// Fallback to older format if RFC3339 fails
		createdAt, parseErr = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if parseErr != nil {
			log.Printf("Warning: Could not parse created_at timestamp %s: %v", createdAtStr, parseErr)
		}
	}
	subscriber.CreatedAt = createdAt

	updatedAt, parseErr := time.Parse(time.RFC3339, updatedAtStr)
	if parseErr != nil {
		// Fallback to older format if RFC3339 fails
		updatedAt, parseErr = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if parseErr != nil {
			log.Printf("Warning: Could not parse updated_at timestamp %s: %v", updatedAtStr, parseErr)
		}
	}
	subscriber.UpdatedAt = updatedAt

	log.Printf("Found subscriber: %s, status: %s", email, subscriber.Status)
	return &subscriber, nil
}

// UpdateStatus updates the status of a subscriber by email.
func (s *SQLiteDB) UpdateStatus(email, status string) error {
	log.Printf("Updating status for %s to: %s", email, status)
	query := `UPDATE subscribers SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE email = ?`

	result, err := s.db.Exec(query, status, email)
	if err != nil {
		log.Printf("ERROR: Failed to update status for %s: %v", email, err)
		return err
	}

	rows, _ := result.RowsAffected()
	log.Printf("Status update for %s successful: %d rows affected", email, rows)
	return nil
}

// ListSubscribers retrieves all subscribers, optionally filtered by status.
// If status is empty, returns all subscribers.
func (s *SQLiteDB) ListSubscribers(status string) ([]*models.Subscriber, error) {
	log.Printf("Listing subscribers with status filter: %s", status)
	var query string
	var rows *sql.Rows
	var err error

	if status == "" {
		query = `SELECT id, email, status, created_at, updated_at FROM subscribers ORDER BY created_at DESC`
		rows, err = s.db.Query(query)
	} else {
		query = `SELECT id, email, status, created_at, updated_at FROM subscribers WHERE status = ? ORDER BY created_at DESC`
		rows, err = s.db.Query(query, status)
	}

	if err != nil {
		log.Printf("ERROR: Failed to query subscribers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var subscribers []*models.Subscriber
	for rows.Next() {
		var subscriber models.Subscriber
		var createdAtStr, updatedAtStr string

		err := rows.Scan(&subscriber.ID, &subscriber.Email, &subscriber.Status, &createdAtStr, &updatedAtStr)
		if err != nil {
			log.Printf("ERROR: Failed to scan subscriber row: %v", err)
			continue
		}

		// Parse timestamps
		createdAt, parseErr := time.Parse(time.RFC3339, createdAtStr)
		if parseErr != nil {
			createdAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		subscriber.CreatedAt = createdAt

		updatedAt, parseErr := time.Parse(time.RFC3339, updatedAtStr)
		if parseErr != nil {
			updatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		}
		subscriber.UpdatedAt = updatedAt

		subscribers = append(subscribers, &subscriber)
	}

	log.Printf("Found %d subscribers with status filter: %s", len(subscribers), status)
	return subscribers, nil
}

// RemoveSubscriber removes a subscriber by email.
func (s *SQLiteDB) RemoveSubscriber(email string) error {
	log.Printf("Removing subscriber: %s", email)
	query := `DELETE FROM subscribers WHERE email = ?`

	result, err := s.db.Exec(query, email)
	if err != nil {
		log.Printf("ERROR: Failed to remove subscriber %s: %v", email, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("No subscriber found with email: %s", email)
		return sql.ErrNoRows
	}

	log.Printf("Successfully removed subscriber %s", email)
	return nil
}

// GetSubscriberStats returns statistics about subscribers by status.
func (s *SQLiteDB) GetSubscriberStats() (map[string]int, error) {
	log.Println("Getting subscriber statistics")
	query := `SELECT status, COUNT(*) as count FROM subscribers GROUP BY status`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Printf("ERROR: Failed to query subscriber stats: %v", err)
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err = rows.Scan(&status, &count)
		if err != nil {
			log.Printf("ERROR: Failed to scan stats row: %v", err)
			continue
		}
		stats[status] = count
	}

	// Add total count
	var totalCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM subscribers`).Scan(&totalCount)
	if err != nil {
		log.Printf("ERROR: Failed to get total count: %v", err)
	} else {
		stats["total"] = totalCount
	}

	log.Printf("Subscriber statistics: %v", stats)
	return stats, nil
}

// HealthCheck performs a comprehensive health check of the database
func (s *SQLiteDB) HealthCheck() error {
	// Test basic connectivity
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := s.db.QueryRow("SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	// Check if required tables exist
	tables := []string{"submissions", "subscribers", "migrations"}
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", table)
		if err := s.db.QueryRow(query).Scan(&count); err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if count == 0 {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}

	log.Println("Database health check passed")
	return nil
}

// GetConnectionStats returns database connection statistics
func (s *SQLiteDB) GetConnectionStats() map[string]interface{} {
	stats := s.db.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// GetRepository returns a repository instance for this database
func (s *SQLiteDB) GetRepository() *Repository {
	return NewRepository(s)
}
