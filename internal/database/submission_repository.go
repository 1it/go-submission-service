package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/1it/go-submission-service/internal/models"
	"github.com/google/uuid"
)

// SQLiteSubmissionRepository implements SubmissionRepository for SQLite
type SQLiteSubmissionRepository struct {
	db *SQLiteDB
}

// Create adds a new submission to the database
func (r *SQLiteSubmissionRepository) Create(submission *models.Submission) error {
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

	_, err = r.db.db.Exec(query,
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

	log.Printf("Created submission with ID: %s", submission.ID)
	return nil
}

// GetByID retrieves a submission by ID
func (r *SQLiteSubmissionRepository) GetByID(id string) (*models.Submission, error) {
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

	err := r.db.db.QueryRow(query, id).Scan(
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
			return nil, fmt.Errorf("submission not found: %s", id)
		}
		return nil, fmt.Errorf("failed to query submission: %w", err)
	}

	// Parse form data JSON
	if err := json.Unmarshal(formDataJSON, &submission.FormData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal form data: %w", err)
	}

	// Parse timestamps and nullable fields
	r.parseSubmissionFields(&submission, createdAtStr, updatedAtStr, processedAtStr, formType, source, ipAddress, userAgent, referrer, sessionID)

	return &submission, nil
}

// Update updates an existing submission
func (r *SQLiteSubmissionRepository) Update(submission *models.Submission) error {
	// Convert form data to JSON
	formDataJSON, err := json.Marshal(submission.FormData)
	if err != nil {
		return fmt.Errorf("failed to marshal form data: %w", err)
	}

	query := `
    UPDATE submissions
    SET form_data = ?, status = ?, form_type = ?, source = ?, ip_address = ?, 
        user_agent = ?, referrer = ?, session_id = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := r.db.db.Exec(query,
		formDataJSON,
		submission.Status,
		submission.FormType,
		submission.Source,
		submission.IPAddress,
		submission.UserAgent,
		submission.Referrer,
		submission.SessionID,
		submission.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("submission not found: %s", submission.ID)
	}

	log.Printf("Updated submission with ID: %s", submission.ID)
	return nil
}

// Delete removes a submission by ID
func (r *SQLiteSubmissionRepository) Delete(id string) error {
	query := `DELETE FROM submissions WHERE id = ?`

	result, err := r.db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete submission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("submission not found: %s", id)
	}

	log.Printf("Deleted submission with ID: %s", id)
	return nil
}

// List retrieves submissions with optional filters
func (r *SQLiteSubmissionRepository) List(filters SubmissionFilters) ([]*models.Submission, error) {
	query, args := r.buildListQuery(filters)

	rows, err := r.db.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()

	return r.scanSubmissions(rows)
}

// Count returns the number of submissions matching the filters
func (r *SQLiteSubmissionRepository) Count(filters SubmissionFilters) (int, error) {
	query, args := r.buildCountQuery(filters)

	var count int
	err := r.db.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count submissions: %w", err)
	}

	return count, nil
}

// GetByEmail retrieves a submission by email field
func (r *SQLiteSubmissionRepository) GetByEmail(email string, emailField string) (*models.Submission, error) {
	// For SQLite with JSON storage, we need to get all submissions and filter
	// In a production system, you might want to use JSON extraction functions
	submissions, err := r.List(SubmissionFilters{})
	if err != nil {
		return nil, err
	}

	for _, submission := range submissions {
		if emailValue, ok := submission.FormData[emailField].(string); ok && emailValue == email {
			return submission, nil
		}
	}

	return nil, fmt.Errorf("submission not found for email: %s", email)
}

// UpdateStatus updates the status of a submission
func (r *SQLiteSubmissionRepository) UpdateStatus(id string, status string) error {
	query := `
    UPDATE submissions
    SET status = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := r.db.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update submission status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("submission not found: %s", id)
	}

	log.Printf("Updated status for submission %s to: %s", id, status)
	return nil
}

// MarkProcessed marks a submission as processed
func (r *SQLiteSubmissionRepository) MarkProcessed(id string) error {
	query := `
    UPDATE submissions
    SET status = 'processed', processed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
    `

	result, err := r.db.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark submission as processed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("submission not found: %s", id)
	}

	log.Printf("Marked submission %s as processed", id)
	return nil
}

// GetByStatus retrieves submissions by status
func (r *SQLiteSubmissionRepository) GetByStatus(status string) ([]*models.Submission, error) {
	return r.List(SubmissionFilters{Status: status})
}

// GetByFormType retrieves submissions by form type
func (r *SQLiteSubmissionRepository) GetByFormType(formType string) ([]*models.Submission, error) {
	return r.List(SubmissionFilters{FormType: formType})
}

// GetFormTypes returns all unique form types
func (r *SQLiteSubmissionRepository) GetFormTypes() ([]string, error) {
	query := `SELECT DISTINCT form_type FROM submissions WHERE form_type IS NOT NULL ORDER BY form_type`

	rows, err := r.db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query form types: %w", err)
	}
	defer rows.Close()

	var formTypes []string
	for rows.Next() {
		var formType string
		if err := rows.Scan(&formType); err != nil {
			log.Printf("ERROR: Failed to scan form type: %v", err)
			continue
		}
		formTypes = append(formTypes, formType)
	}

	return formTypes, nil
}

// GetByDateRange retrieves submissions within a date range
func (r *SQLiteSubmissionRepository) GetByDateRange(startDate, endDate time.Time) ([]*models.Submission, error) {
	return r.List(SubmissionFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
}

// GetStats returns statistics about submissions by status
func (r *SQLiteSubmissionRepository) GetStats() (map[string]int, error) {
	query := `SELECT status, COUNT(*) as count FROM submissions GROUP BY status`

	rows, err := r.db.db.Query(query)
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
	err = r.db.db.QueryRow(`SELECT COUNT(*) FROM submissions`).Scan(&totalCount)
	if err != nil {
		log.Printf("ERROR: Failed to get total count: %v", err)
	} else {
		stats["total"] = totalCount
	}

	return stats, nil
}

// GetStatsByFormType returns statistics grouped by form type
func (r *SQLiteSubmissionRepository) GetStatsByFormType() (map[string]map[string]int, error) {
	query := `
    SELECT form_type, status, COUNT(*) as count 
    FROM submissions 
    GROUP BY form_type, status
    ORDER BY form_type, status
    `

	rows, err := r.db.db.Query(query)
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

// GetStatsByDateRange returns statistics for a date range
func (r *SQLiteSubmissionRepository) GetStatsByDateRange(startDate, endDate time.Time) (map[string]int, error) {
	query := `
    SELECT status, COUNT(*) as count 
    FROM submissions 
    WHERE created_at BETWEEN ? AND ?
    GROUP BY status
    `

	rows, err := r.db.db.Query(query, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to query submission stats by date range: %w", err)
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

	// Add total count for the date range
	var totalCount int
	err = r.db.db.QueryRow(`
		SELECT COUNT(*) FROM submissions 
		WHERE created_at BETWEEN ? AND ?
	`, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339)).Scan(&totalCount)
	if err != nil {
		log.Printf("ERROR: Failed to get total count for date range: %v", err)
	} else {
		stats["total"] = totalCount
	}

	return stats, nil
}

// HealthCheck verifies the database connection is working
func (r *SQLiteSubmissionRepository) HealthCheck() error {
	query := `SELECT 1`
	var result int
	err := r.db.db.QueryRow(query).Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Helper methods

// buildListQuery constructs the SQL query for listing submissions with filters
func (r *SQLiteSubmissionRepository) buildListQuery(filters SubmissionFilters) (string, []interface{}) {
	baseQuery := `
    SELECT id, form_data, status, form_type, source, ip_address, user_agent, 
           referrer, session_id, created_at, updated_at, processed_at
    FROM submissions
    `

	var conditions []string
	var args []interface{}

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.FormType != "" {
		conditions = append(conditions, "form_type = ?")
		args = append(args, filters.FormType)
	}

	if filters.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filters.Source)
	}

	if filters.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filters.IPAddress)
	}

	if filters.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, filters.SessionID)
	}

	if filters.StartDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filters.StartDate.Format(time.RFC3339))
	}

	if filters.EndDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filters.EndDate.Format(time.RFC3339))
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	orderBy := "created_at"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	orderDir := "DESC"
	if filters.OrderDir != "" {
		orderDir = filters.OrderDir
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Add pagination
	if filters.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, filters.Limit)

		if filters.Offset > 0 {
			baseQuery += " OFFSET ?"
			args = append(args, filters.Offset)
		}
	}

	return baseQuery, args
}

// buildCountQuery constructs the SQL query for counting submissions with filters
func (r *SQLiteSubmissionRepository) buildCountQuery(filters SubmissionFilters) (string, []interface{}) {
	baseQuery := `SELECT COUNT(*) FROM submissions`

	var conditions []string
	var args []interface{}

	if filters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filters.Status)
	}

	if filters.FormType != "" {
		conditions = append(conditions, "form_type = ?")
		args = append(args, filters.FormType)
	}

	if filters.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filters.Source)
	}

	if filters.IPAddress != "" {
		conditions = append(conditions, "ip_address = ?")
		args = append(args, filters.IPAddress)
	}

	if filters.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, filters.SessionID)
	}

	if filters.StartDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filters.StartDate.Format(time.RFC3339))
	}

	if filters.EndDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filters.EndDate.Format(time.RFC3339))
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	return baseQuery, args
}

// scanSubmissions is a helper method to scan submission rows
func (r *SQLiteSubmissionRepository) scanSubmissions(rows *sql.Rows) ([]*models.Submission, error) {
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
		if err := json.Unmarshal(formDataJSON, &submission.FormData); err != nil {
			log.Printf("ERROR: Failed to unmarshal form data: %v", err)
			continue
		}

		// Parse timestamps and nullable fields
		r.parseSubmissionFields(&submission, createdAtStr, updatedAtStr, processedAtStr, formType, source, ipAddress, userAgent, referrer, sessionID)

		submissions = append(submissions, &submission)
	}

	return submissions, nil
}

// parseSubmissionFields parses timestamps and nullable fields for a submission
func (r *SQLiteSubmissionRepository) parseSubmissionFields(submission *models.Submission, createdAtStr, updatedAtStr string, processedAtStr sql.NullString, formType, source, ipAddress, userAgent, referrer, sessionID sql.NullString) {
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
}
