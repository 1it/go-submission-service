package database

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version   int       `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

// GetMigrations returns all available migrations in order
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create initial subscribers table",
			Up: `
				CREATE TABLE IF NOT EXISTS subscribers (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					email TEXT NOT NULL UNIQUE,
					status TEXT DEFAULT 'pending',
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);
			`,
			Down: `DROP TABLE IF EXISTS subscribers;`,
		},
		{
			Version:     2,
			Description: "Create submissions table with flexible JSON data",
			Up: `
				CREATE TABLE IF NOT EXISTS submissions (
					id TEXT PRIMARY KEY,
					form_data TEXT NOT NULL,
					status TEXT DEFAULT 'pending',
					ip_address TEXT,
					user_agent TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);
			`,
			Down: `DROP TABLE IF EXISTS submissions;`,
		},
		{
			Version:     3,
			Description: "Add indexes for better performance",
			Up: `
				CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status);
				CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON submissions(created_at);
				CREATE INDEX IF NOT EXISTS idx_subscribers_email ON subscribers(email);
				CREATE INDEX IF NOT EXISTS idx_subscribers_status ON subscribers(status);
			`,
			Down: `
				DROP INDEX IF EXISTS idx_submissions_status;
				DROP INDEX IF EXISTS idx_submissions_created_at;
				DROP INDEX IF EXISTS idx_subscribers_email;
				DROP INDEX IF EXISTS idx_subscribers_status;
			`,
		},
		{
			Version:     4,
			Description: "Add form_type and source fields to submissions",
			Up: `
				ALTER TABLE submissions ADD COLUMN form_type TEXT DEFAULT 'generic';
				ALTER TABLE submissions ADD COLUMN source TEXT;
				CREATE INDEX IF NOT EXISTS idx_submissions_form_type ON submissions(form_type);
			`,
			Down: `
				-- SQLite doesn't support DROP COLUMN, so we'd need to recreate the table
				-- For now, we'll leave the columns as they won't hurt
				DROP INDEX IF EXISTS idx_submissions_form_type;
			`,
		},
		{
			Version:     5,
			Description: "Add additional metadata fields to submissions",
			Up: `
				ALTER TABLE submissions ADD COLUMN referrer TEXT;
				ALTER TABLE submissions ADD COLUMN session_id TEXT;
				ALTER TABLE submissions ADD COLUMN processed_at TIMESTAMP;
				CREATE INDEX IF NOT EXISTS idx_submissions_session_id ON submissions(session_id);
				CREATE INDEX IF NOT EXISTS idx_submissions_processed_at ON submissions(processed_at);
			`,
			Down: `
				-- SQLite doesn't support DROP COLUMN, so we'd need to recreate the table
				-- For now, we'll leave the columns as they won't hurt
				DROP INDEX IF EXISTS idx_submissions_session_id;
				DROP INDEX IF EXISTS idx_submissions_processed_at;
			`,
		},
	}
}

// InitializeMigrationTable creates the migrations table if it doesn't exist
func (db *SQLiteDB) InitializeMigrationTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// GetAppliedMigrations returns all applied migrations
func (db *SQLiteDB) GetAppliedMigrations() ([]MigrationRecord, error) {
	query := `SELECT version, applied_at FROM migrations ORDER BY version`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var migration MigrationRecord
		var appliedAtStr string

		err := rows.Scan(&migration.Version, &appliedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration record: %w", err)
		}

		// Parse timestamp
		appliedAt, err := time.Parse(time.RFC3339, appliedAtStr)
		if err != nil {
			appliedAt, _ = time.Parse("2006-01-02 15:04:05", appliedAtStr)
		}
		migration.AppliedAt = appliedAt

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// ApplyMigration applies a single migration
func (db *SQLiteDB) ApplyMigration(migration Migration) error {
	log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

	// Start transaction
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	statements := strings.Split(migration.Up, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute migration statement: %w", err)
		}
	}

	// Record migration as applied
	_, err = tx.Exec(
		`INSERT INTO migrations (version, description) VALUES (?, ?)`,
		migration.Version, migration.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Printf("Successfully applied migration %d", migration.Version)
	return nil
}

// RollbackMigration rolls back a single migration
func (db *SQLiteDB) RollbackMigration(migration Migration) error {
	log.Printf("Rolling back migration %d: %s", migration.Version, migration.Description)

	// Start transaction
	tx, err := db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	statements := strings.Split(migration.Down, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err = tx.Exec(stmt)
		if err != nil {
			return fmt.Errorf("failed to execute rollback statement: %w", err)
		}
	}

	// Remove migration record
	_, err = tx.Exec(`DELETE FROM migrations WHERE version = ?`, migration.Version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("Successfully rolled back migration %d", migration.Version)
	return nil
}

// Migrate runs all pending migrations
func (db *SQLiteDB) Migrate() error {
	log.Println("Starting database migration")

	// Initialize migration table
	if err := db.InitializeMigrationTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := db.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create map of applied versions for quick lookup
	appliedVersions := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Get all available migrations
	allMigrations := GetMigrations()

	// Sort migrations by version
	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	// Apply pending migrations
	pendingCount := 0
	for _, migration := range allMigrations {
		if !appliedVersions[migration.Version] {
			if err := db.ApplyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
			pendingCount++
		}
	}

	if pendingCount == 0 {
		log.Println("No pending migrations to apply")
	} else {
		log.Printf("Successfully applied %d migrations", pendingCount)
	}

	return nil
}

// MigrateToVersion migrates to a specific version (up or down)
func (db *SQLiteDB) MigrateToVersion(targetVersion int) error {
	log.Printf("Migrating to version %d", targetVersion)

	// Initialize migration table
	if err := db.InitializeMigrationTable(); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := db.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get current version
	currentVersion := 0
	if len(appliedMigrations) > 0 {
		currentVersion = appliedMigrations[len(appliedMigrations)-1].Version
	}

	if currentVersion == targetVersion {
		log.Printf("Already at version %d", targetVersion)
		return nil
	}

	allMigrations := GetMigrations()

	if targetVersion > currentVersion {
		// Migrate up
		for _, migration := range allMigrations {
			if migration.Version > currentVersion && migration.Version <= targetVersion {
				if err := db.ApplyMigration(migration); err != nil {
					return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
				}
			}
		}
	} else {
		// Migrate down
		// Sort migrations in reverse order for rollback
		sort.Slice(allMigrations, func(i, j int) bool {
			return allMigrations[i].Version > allMigrations[j].Version
		})

		for _, migration := range allMigrations {
			if migration.Version <= currentVersion && migration.Version > targetVersion {
				if err := db.RollbackMigration(migration); err != nil {
					return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
				}
			}
		}
	}

	log.Printf("Successfully migrated to version %d", targetVersion)
	return nil
}

// GetMigrationStatus returns the current migration status
func (db *SQLiteDB) GetMigrationStatus() (map[string]interface{}, error) {
	appliedMigrations, err := db.GetAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	allMigrations := GetMigrations()

	currentVersion := 0
	if len(appliedMigrations) > 0 {
		currentVersion = appliedMigrations[len(appliedMigrations)-1].Version
	}

	latestVersion := 0
	if len(allMigrations) > 0 {
		for _, migration := range allMigrations {
			if migration.Version > latestVersion {
				latestVersion = migration.Version
			}
		}
	}

	pendingMigrations := []Migration{}
	appliedVersions := make(map[int]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	for _, migration := range allMigrations {
		if !appliedVersions[migration.Version] {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return map[string]interface{}{
		"current_version":    currentVersion,
		"latest_version":     latestVersion,
		"applied_count":      len(appliedMigrations),
		"pending_count":      len(pendingMigrations),
		"applied_migrations": appliedMigrations,
		"pending_migrations": pendingMigrations,
	}, nil
}
