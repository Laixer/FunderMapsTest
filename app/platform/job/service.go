package job

import (
	"time"

	"gorm.io/gorm"

	"fundermaps/app/config"
	"fundermaps/app/database"
)

// Service handles business logic for worker jobs
type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewService creates a new job service
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		cfg: cfg,
	}
}

// CreateJobInput defines the input structure for creating a new worker job
type CreateJobInput struct {
	JobType      string                 `json:"job_type" validate:"required"`
	Payload      map[string]interface{} `json:"payload"`
	Priority     *int                   `json:"priority"`
	MaxRetries   *int                   `json:"max_retries"`
	ProcessAfter *time.Time             `json:"process_after"`
}

// CreateJob creates a new worker job with the provided input
func (s *Service) CreateJob(input CreateJobInput) (*database.WorkerJob, error) {
	// Create a new job with default values
	job := database.WorkerJob{
		JobType:    input.JobType,
		Payload:    database.JSONObject(input.Payload),
		Status:     database.JobStatusPending,
		RetryCount: 0,
	}

	// Set optional fields if provided
	if input.Priority != nil {
		job.Priority = *input.Priority
	}

	if input.MaxRetries != nil {
		job.MaxRetries = *input.MaxRetries
	}

	if input.ProcessAfter != nil {
		job.ProcessAfter = input.ProcessAfter
	}

	// Create the job in the database
	result := s.db.Create(&job)
	if result.Error != nil {
		return nil, result.Error
	}

	return &job, nil
}

// GetAllJobsOptions defines options for filtering jobs
type GetAllJobsOptions struct {
	JobType string
	Status  string
	Limit   int
	Offset  int
}

// GetAllJobs retrieves a list of worker jobs with optional filtering
func (s *Service) GetAllJobs(options GetAllJobsOptions) ([]database.WorkerJob, error) {
	var jobs []database.WorkerJob

	// Apply query parameters for filtering
	query := s.db

	// Filter by job type if provided
	if options.JobType != "" {
		query = query.Where("job_type = ?", options.JobType)
	}

	// Filter by status if provided
	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}

	// Apply pagination with defaults
	limit := options.Limit
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	offset := options.Offset
	if offset < 0 {
		offset = 0
	}

	// Execute the query
	result := query.Order("priority DESC, created_at ASC").Limit(limit).Offset(offset).Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}

	return jobs, nil
}

// GetJobByID retrieves a specific job by ID
func (s *Service) GetJobByID(jobID string) (*database.WorkerJob, error) {
	var job database.WorkerJob
	result := s.db.First(&job, jobID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &job, nil
}

// CancelJob cancels a job by changing its status to failed
func (s *Service) CancelJob(jobID string) (*database.WorkerJob, error) {
	var job database.WorkerJob
	result := s.db.First(&job, jobID)
	if result.Error != nil {
		return nil, result.Error
	}

	// Can only cancel pending or retry jobs
	if job.Status != database.JobStatusPending && job.Status != database.JobStatusRetry {
		return nil, gorm.ErrInvalidValue // Use a more specific error type in production
	}

	// Update the job status to failed
	cancelReason := "Cancelled by admin"
	job.Status = database.JobStatusFailed
	job.LastError = &cancelReason
	job.UpdatedAt = time.Now()

	result = s.db.Save(&job)
	if result.Error != nil {
		return nil, result.Error
	}

	return &job, nil
}

// UpdateJobStatus updates the status of a job
func (s *Service) UpdateJobStatus(jobID string, status database.JobStatus) error {
	result := s.db.Model(&database.WorkerJob{}).Where("id = ?", jobID).Update("status", status)
	return result.Error
}

// IncrementRetryCount increments the retry count for a job
func (s *Service) IncrementRetryCount(jobID string) error {
	result := s.db.Model(&database.WorkerJob{}).Where("id = ?", jobID).Update("retry_count", gorm.Expr("retry_count + 1"))
	return result.Error
}

// GetPendingJobs retrieves jobs that are ready to be processed
func (s *Service) GetPendingJobs(limit int) ([]database.WorkerJob, error) {
	var jobs []database.WorkerJob

	if limit <= 0 {
		limit = 10
	}

	result := s.db.Where("status IN ? AND (process_after IS NULL OR process_after <= ?)",
		[]database.JobStatus{database.JobStatusPending, database.JobStatusRetry},
		time.Now()).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&jobs)

	if result.Error != nil {
		return nil, result.Error
	}

	return jobs, nil
}

// MarkJobAsRunning marks a job as currently running
func (s *Service) MarkJobAsRunning(jobID string) error {
	result := s.db.Model(&database.WorkerJob{}).
		Where("id = ? AND status IN ?", jobID, []database.JobStatus{database.JobStatusPending, database.JobStatusRetry}).
		Updates(map[string]interface{}{
			"status":     database.JobStatusRunning,
			"updated_at": time.Now(),
		})
	return result.Error
}

// MarkJobAsComplete marks a job as successfully completed
func (s *Service) MarkJobAsComplete(jobID string) error {
	result := s.db.Model(&database.WorkerJob{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":     database.JobStatusComplete,
			"updated_at": time.Now(),
		})
	return result.Error
}

// MarkJobAsFailed marks a job as failed with an error message
func (s *Service) MarkJobAsFailed(jobID string, errorMsg string) error {
	result := s.db.Model(&database.WorkerJob{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":     database.JobStatusFailed,
			"last_error": errorMsg,
			"updated_at": time.Now(),
		})
	return result.Error
}
