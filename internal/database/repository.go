package database

import (
	"time"

	"github.com/1it/go-submission-service/internal/models"
)

// SubmissionRepository defines the interface for submission data operations
type SubmissionRepository interface {
	// Basic CRUD operations
	Create(submission *models.Submission) error
	GetByID(id string) (*models.Submission, error)
	Update(submission *models.Submission) error
	Delete(id string) error

	// Query operations
	List(filters SubmissionFilters) ([]*models.Submission, error)
	Count(filters SubmissionFilters) (int, error)
	GetByEmail(email string, emailField string) (*models.Submission, error)

	// Status operations
	UpdateStatus(id string, status string) error
	MarkProcessed(id string) error
	GetByStatus(status string) ([]*models.Submission, error)

	// Form type operations
	GetByFormType(formType string) ([]*models.Submission, error)
	GetFormTypes() ([]string, error)

	// Date range operations
	GetByDateRange(startDate, endDate time.Time) ([]*models.Submission, error)

	// Statistics
	GetStats() (map[string]int, error)
	GetStatsByFormType() (map[string]map[string]int, error)
	GetStatsByDateRange(startDate, endDate time.Time) (map[string]int, error)

	// Health check
	HealthCheck() error
}

// SubscriberRepository defines the interface for subscriber data operations (legacy support)
type SubscriberRepository interface {
	// Basic CRUD operations
	Add(email string) error
	GetByEmail(email string) (*models.Subscriber, error)
	UpdateStatus(email, status string) error
	Remove(email string) error

	// Query operations
	List(status string) ([]*models.Subscriber, error)
	GetStats() (map[string]int, error)

	// Health check
	HealthCheck() error
}

// SubmissionFilters defines filters for querying submissions
type SubmissionFilters struct {
	Status    string
	FormType  string
	Source    string
	IPAddress string
	SessionID string
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
	OrderBy   string
	OrderDir  string // "ASC" or "DESC"
}

// Repository aggregates all repository interfaces
type Repository struct {
	Submissions SubmissionRepository
	Subscribers SubscriberRepository
}

// NewRepository creates a new repository instance
func NewRepository(db *SQLiteDB) *Repository {
	return &Repository{
		Submissions: &SQLiteSubmissionRepository{db: db},
		Subscribers: &SQLiteSubscriberRepository{db: db},
	}
}
