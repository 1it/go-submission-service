package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/1it/go-submission-service/internal/models"
)

// SQLiteSubscriberRepository implements SubscriberRepository for SQLite
type SQLiteSubscriberRepository struct {
	db *SQLiteDB
}

// Add adds a new subscriber or updates the timestamp if it exists
func (r *SQLiteSubscriberRepository) Add(email string) error {
	log.Printf("Adding or updating subscriber: %s", email)
	query := `
    INSERT INTO subscribers (email) 
    VALUES (?) 
    ON CONFLICT(email) DO UPDATE SET 
        updated_at = CURRENT_TIMESTAMP
    `

	result, err := r.db.db.Exec(query, email)
	if err != nil {
		log.Printf("ERROR: Failed to add/update subscriber %s: %v", email, err)
		return fmt.Errorf("failed to add subscriber: %w", err)
	}

	rows, _ := result.RowsAffected()
	log.Printf("Subscriber %s operation successful: %d rows affected", email, rows)
	return nil
}

// GetByEmail retrieves a subscriber by email
func (r *SQLiteSubscriberRepository) GetByEmail(email string) (*models.Subscriber, error) {
	log.Printf("Looking up subscriber by email: %s", email)
	query := `SELECT id, email, status, created_at, updated_at FROM subscribers WHERE email = ?`

	row := r.db.db.QueryRow(query, email)

	var subscriber models.Subscriber
	var createdAtStr, updatedAtStr string

	err := row.Scan(&subscriber.ID, &subscriber.Email, &subscriber.Status, &createdAtStr, &updatedAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Subscriber not found: %s", email)
			return nil, fmt.Errorf("subscriber not found: %s", email)
		} else {
			log.Printf("ERROR: Failed to retrieve subscriber %s: %v", email, err)
			return nil, fmt.Errorf("failed to retrieve subscriber: %w", err)
		}
	}

	// Parse timestamps
	createdAt, parseErr := time.Parse(time.RFC3339, createdAtStr)
	if parseErr != nil {
		createdAt, parseErr = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if parseErr != nil {
			log.Printf("Warning: Could not parse created_at timestamp %s: %v", createdAtStr, parseErr)
		}
	}
	subscriber.CreatedAt = createdAt

	updatedAt, parseErr := time.Parse(time.RFC3339, updatedAtStr)
	if parseErr != nil {
		updatedAt, parseErr = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if parseErr != nil {
			log.Printf("Warning: Could not parse updated_at timestamp %s: %v", updatedAtStr, parseErr)
		}
	}
	subscriber.UpdatedAt = updatedAt

	log.Printf("Found subscriber: %s, status: %s", email, subscriber.Status)
	return &subscriber, nil
}

// UpdateStatus updates the status of a subscriber by email
func (r *SQLiteSubscriberRepository) UpdateStatus(email, status string) error {
	log.Printf("Updating status for %s to: %s", email, status)
	query := `UPDATE subscribers SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE email = ?`

	result, err := r.db.db.Exec(query, status, email)
	if err != nil {
		log.Printf("ERROR: Failed to update status for %s: %v", email, err)
		return fmt.Errorf("failed to update subscriber status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("subscriber not found: %s", email)
	}

	log.Printf("Status update for %s successful: %d rows affected", email, rows)
	return nil
}

// Remove removes a subscriber by email
func (r *SQLiteSubscriberRepository) Remove(email string) error {
	log.Printf("Removing subscriber: %s", email)
	query := `DELETE FROM subscribers WHERE email = ?`

	result, err := r.db.db.Exec(query, email)
	if err != nil {
		log.Printf("ERROR: Failed to remove subscriber %s: %v", email, err)
		return fmt.Errorf("failed to remove subscriber: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		log.Printf("No subscriber found with email: %s", email)
		return fmt.Errorf("subscriber not found: %s", email)
	}

	log.Printf("Successfully removed subscriber %s", email)
	return nil
}

// List retrieves all subscribers, optionally filtered by status
func (r *SQLiteSubscriberRepository) List(status string) ([]*models.Subscriber, error) {
	log.Printf("Listing subscribers with status filter: %s", status)
	var query string
	var rows *sql.Rows
	var err error

	if status == "" {
		query = `SELECT id, email, status, created_at, updated_at FROM subscribers ORDER BY created_at DESC`
		rows, err = r.db.db.Query(query)
	} else {
		query = `SELECT id, email, status, created_at, updated_at FROM subscribers WHERE status = ? ORDER BY created_at DESC`
		rows, err = r.db.db.Query(query, status)
	}

	if err != nil {
		log.Printf("ERROR: Failed to query subscribers: %v", err)
		return nil, fmt.Errorf("failed to query subscribers: %w", err)
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

// GetStats returns statistics about subscribers by status
func (r *SQLiteSubscriberRepository) GetStats() (map[string]int, error) {
	log.Println("Getting subscriber statistics")
	query := `SELECT status, COUNT(*) as count FROM subscribers GROUP BY status`

	rows, err := r.db.db.Query(query)
	if err != nil {
		log.Printf("ERROR: Failed to query subscriber stats: %v", err)
		return nil, fmt.Errorf("failed to query subscriber stats: %w", err)
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
	err = r.db.db.QueryRow(`SELECT COUNT(*) FROM subscribers`).Scan(&totalCount)
	if err != nil {
		log.Printf("ERROR: Failed to get total count: %v", err)
	} else {
		stats["total"] = totalCount
	}

	log.Printf("Subscriber statistics: %v", stats)
	return stats, nil
}

// HealthCheck verifies the database connection is working
func (r *SQLiteSubscriberRepository) HealthCheck() error {
	query := `SELECT 1`
	var result int
	err := r.db.db.QueryRow(query).Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
