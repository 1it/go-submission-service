package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/1it/go-submission-service/internal/models"
	"github.com/google/uuid"
)

// AddSubmission adds a new submission to the database
func (db *SQLiteDB) AddSubmission(formData map[string]interface{}) (string, error) {
	// Generate a unique ID for the submission
	id := uuid.New().String()

	// Convert form data to JSON
	formDataJSON, err := json.Marshal(formData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal form data: %w", err)
	}

	// Insert into database with basic fields
	query := `
    INSERT INTO submissions (id, form_data, status, form_type, created_at, updated_at)
    VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	// Extract form type from form data if available, default to 'generic'
	formType := "generic"
	if ft, ok := formData["form_type"].(string); ok && ft != "" {
		formType = ft
	}

	_, err = db.db.Exec(query, id, formDataJSON, "pending", formType)
	if err != nil {
		return "", fmt.Errorf("failed to insert submission: %w", err)
	}

	return id, nil
}

// AddSubmissionWithMetadata adds a new submission with additional metadata
func (db *SQLiteDB) AddSubmissionWithMetadata(submission *models.Submission) error {
	// Generate a unique ID if not provided
	if submission.ID == "" {
		submission.ID = uuid.New().String()
	}

	// Convert form data to JSON
	formDataJSON, err := json.Marshal(submission.FormData)
	if err != nil {
		return fmt.Errorf("failed to marshal form data: %w", err)
	}

	// Set defaults
	if submission.Status == "" {
		submission.Status = "pending"
	}
	if submission.FormType == "" {
		submission.FormType = "generic"
	}

	// Insert into database with all metadata
	query := `
    INSERT INTO submissions (
        id, form_data, status, form_type, source, ip_address, user_agent, 
        referrer, session_id, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	_, err = db.db.Exec(query,
		submission.ID,
		formDataJSON,
		submission.Status,
		submission.FormType,
		submission.Source,
		submission.IPAddress,
		submission.UserAgent,
		submission.Referrer,
		submission.SessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert submission: %w", err)
	}

	return nil
}

// GetSubmission retrieves a submission by ID
func (db *SQLiteDB) GetSubmission(id string) (*models.Submission, error) {
	query := `
    SELECT id, form_data, status, form_type, source, ip_address, user_agent, 
           referrer, session_id, created_at, updated_at, processed_at
    FROM submissions
    WHERE id = ?
    `

	var submission models.Submission
	var formDataJSON []byte
	var createdAtStr, updatedAtStr string
	var formType, source, ipAddress, userAgent, referrer, sessionID sql.NullString
	var processedAtStr sql.NullString

	err := db.db.QueryRow(query, id).Scan(
		&submission.ID,
		&formDataJSON,
		&submission.Status,
		&formType,
		&source,
		&ipAddress,
		&userAgent,
		&referrer,
		&sessionID,
		&createdAtStr,
		&updatedAtStr,
		&processedAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to query submission: %w", err)
	}

	// Parse form data JSON
	if err = json.Unmarshal(formDataJSON, &submission.FormData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal form data: %w", err)
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		createdAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	}
	submission.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		updatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)
	}
	submission.UpdatedAt = updatedAt

	// Handle nullable fields
	if formType.Valid {
		submission.FormType = formType.String
	}
	if source.Valid {
		submission.Source = source.String
	}
	if ipAddress.Valid {
		submission.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		submission.UserAgent = userAgent.String
	}
	if referrer.Valid {
		submission.Referrer = referrer.String
	}
	if sessionID.Valid {
		submission.SessionID = sessionID.String
	}
	if processedAtStr.Valid {
		processedAt, err := time.Parse(time.RFC3339, processedAtStr.String)
		if err != nil {
			processedAt, _ = time.Parse("2006-01-02 15:04:05", processedAtStr.String)
		}
		submission.ProcessedAt = &processedAt
	}

	return &submission, nil
}

// GetSubmissionByEmail retrieves a submission by email field
func (db *SQLiteDB) GetSubmissionByEmail(email string, emailField string) (*models.Submission, error) {
	// This is a bit tricky with SQLite since we're storing JSON
	// We'll use a simple approach for now - get all submissions and filter
	submissions, err := db.ListSubmissions("")
	if err != nil {
		return nil, err
	}

	for _, submission := range submissions {
		if emailValue, ok := submission.FormData[emailField].(string); ok && emailValue == email {
			return submission, nil
		}
	}

	return nil, sql.ErrNoRows
}

// UpdateSubmissionStatus updates the status of a submission
func (db *SQLiteDB) UpdateSubmissionStatus(id string, status string) error {
	query := `
    UPDATE submissions
    SET status = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := db.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ListSubmissions retrieves all submissions, optionally filtered by status
func (db *SQLiteDB) ListSubmissions(status string) ([]*models.Submission, error) {
	rows, err := db.querySubmissionsByStatus(status)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return db.scanSubmissions(rows)
}

func (db *SQLiteDB) querySubmissionsByStatus(status string) (*sql.Rows, error) {
	baseQuery := `
        SELECT id, form_data, status, form_type, source, ip_address, user_agent,
               referrer, session_id, created_at, updated_at, processed_at
        FROM submissions
        ORDER BY created_at DESC
        `
	if status == "" {
		return db.db.Query(baseQuery)
	}
	query := `
        SELECT id, form_data, status, form_type, source, ip_address, user_agent,
               referrer, session_id, created_at, updated_at, processed_at
        FROM submissions
        WHERE status = ?
        ORDER BY created_at DESC
        `
	return db.db.Query(query, status)
}

// RemoveSubmission removes a submission by ID
func (db *SQLiteDB) RemoveSubmission(id string) error {
	query := `DELETE FROM submissions WHERE id = ?`

	result, err := db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove submission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetSubmissionStats returns statistics about submissions by status
func (db *SQLiteDB) GetSubmissionStats() (map[string]int, error) {
	query := `SELECT status, COUNT(*) as count FROM submissions GROUP BY status`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query submission stats: %w", err)
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
	err = db.db.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&totalCount)
	if err != nil {
		log.Printf("ERROR: Failed to get total count: %v", err)
	} else {
		stats["total"] = totalCount
	}

	return stats, nil
}

// ListSubmissionsByFormType retrieves submissions filtered by form type
func (db *SQLiteDB) ListSubmissionsByFormType(formType string) ([]*models.Submission, error) {
	query := `
    SELECT id, form_data, status, form_type, source, ip_address, user_agent, 
           referrer, session_id, created_at, updated_at, processed_at
    FROM submissions
    WHERE form_type = ?
    ORDER BY created_at DESC
    `

	rows, err := db.db.Query(query, formType)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions by form type: %w", err)
	}
	defer rows.Close()

	return db.scanSubmissions(rows)
}

// ListSubmissionsByDateRange retrieves submissions within a date range
func (db *SQLiteDB) ListSubmissionsByDateRange(startDate, endDate time.Time) ([]*models.Submission, error) {
	query := `
    SELECT id, form_data, status, form_type, source, ip_address, user_agent, 
           referrer, session_id, created_at, updated_at, processed_at
    FROM submissions
    WHERE created_at BETWEEN ? AND ?
    ORDER BY created_at DESC
    `

	rows, err := db.db.Query(query, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions by date range: %w", err)
	}
	defer rows.Close()

	return db.scanSubmissions(rows)
}

// MarkSubmissionProcessed marks a submission as processed
func (db *SQLiteDB) MarkSubmissionProcessed(id string) error {
	query := `
    UPDATE submissions
    SET status = 'processed', processed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark submission as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetSubmissionStatsByFormType returns statistics grouped by form type
func (db *SQLiteDB) GetSubmissionStatsByFormType() (map[string]map[string]int, error) {
	query := `
    SELECT form_type, status, COUNT(*) as count 
    FROM submissions 
    GROUP BY form_type, status
    ORDER BY form_type, status
    `

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query submission stats by form type: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]map[string]int)
	for rows.Next() {
		var formType, status string
		var count int

		err := rows.Scan(&formType, &status, &count)
		if err != nil {
			log.Printf("ERROR: Failed to scan stats row: %v", err)
			continue
		}

		if stats[formType] == nil {
			stats[formType] = make(map[string]int)
		}
		stats[formType][status] = count
	}

	return stats, nil
}

// scanSubmissions is a helper method to scan submission rows
func (db *SQLiteDB) scanSubmissions(rows *sql.Rows) ([]*models.Submission, error) {
	var submissions []*models.Submission

	for rows.Next() {
		var submission models.Submission
		var formDataJSON []byte
		var createdAtStr, updatedAtStr string
		var formType, source, ipAddress, userAgent, referrer, sessionID sql.NullString
		var processedAtStr sql.NullString

		err := rows.Scan(
			&submission.ID,
			&formDataJSON,
			&submission.Status,
			&formType,
			&source,
			&ipAddress,
			&userAgent,
			&referrer,
			&sessionID,
			&createdAtStr,
			&updatedAtStr,
			&processedAtStr,
		)

		if err != nil {
			log.Printf("ERROR: Failed to scan submission row: %v", err)
			continue
		}

		// Parse form data JSON
		if err = json.Unmarshal(formDataJSON, &submission.FormData); err != nil {
			log.Printf("ERROR: Failed to unmarshal form data: %v", err)
			continue
		}

		// Parse timestamps
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			createdAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		}
		submission.CreatedAt = createdAt

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			updatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		}
		submission.UpdatedAt = updatedAt

		// Handle nullable fields
		if formType.Valid {
			submission.FormType = formType.String
		}
		if source.Valid {
			submission.Source = source.String
		}
		if ipAddress.Valid {
			submission.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			submission.UserAgent = userAgent.String
		}
		if referrer.Valid {
			submission.Referrer = referrer.String
		}
		if sessionID.Valid {
			submission.SessionID = sessionID.String
		}
		if processedAtStr.Valid {
			processedAt, err := time.Parse(time.RFC3339, processedAtStr.String)
			if err != nil {
				processedAt, _ = time.Parse("2006-01-02 15:04:05", processedAtStr.String)
			}
			submission.ProcessedAt = &processedAt
		}

		submissions = append(submissions, &submission)
	}

	return submissions, nil
}
