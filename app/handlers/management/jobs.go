package mngmt

import (
	"time"

	"fundermaps/app/config"
	"fundermaps/app/database"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateJobInput defines the input structure for creating a new worker job
type CreateJobInput struct {
	JobType      string                 `json:"job_type" validate:"required"`
	Payload      map[string]interface{} `json:"payload"`
	Priority     *int                   `json:"priority"`
	MaxRetries   *int                   `json:"max_retries"`
	ProcessAfter *time.Time             `json:"process_after"`
}

// CreateJob handles the creation of a new background worker job
func CreateJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var input CreateJobInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	// Validate the input
	if err := config.Validate.Struct(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

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
	result := db.Create(&job)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create job",
			"error":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(job)
}

// GetAllJobs retrieves a list of worker jobs with optional filtering
func GetAllJobs(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	var jobs []database.WorkerJob

	// Apply query parameters for filtering
	query := db

	// Filter by job type if provided
	if jobType := c.Query("job_type"); jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Pagination
	limit := min(c.QueryInt("limit", 100), 100)
	offset := c.QueryInt("offset", 0)

	// Execute the query
	result := query.Order("priority DESC, created_at ASC").Limit(limit).Offset(offset).Find(&jobs)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(jobs)
}

// GetJob retrieves a specific job by ID
func GetJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	jobID := c.Params("id")

	var job database.WorkerJob
	result := db.First(&job, jobID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(job)
}

// CancelJob cancels a job by changing its status to failed
func CancelJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)

	jobID := c.Params("id")

	var job database.WorkerJob
	result := db.First(&job, jobID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	// Can only cancel pending or retry jobs
	if job.Status != database.JobStatusPending && job.Status != database.JobStatusRetry {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cannot cancel job that is not in pending or retry status",
		})
	}

	// Update the job status to failed
	cancelReason := "Cancelled by admin"
	job.Status = database.JobStatusFailed
	job.LastError = &cancelReason
	job.UpdatedAt = time.Now()

	result = db.Save(&job)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to cancel job",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(job)
}
