package mngmt

import (
	"fundermaps/app/config"
	"fundermaps/app/platform/job"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateJob handles the creation of a new background worker job
func CreateJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	cfg := c.Locals("config").(*config.Config)

	jobService := job.NewService(db, cfg)

	var input job.CreateJobInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
	}

	// Validate the input
	if err := config.Validate.Struct(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// Create the job using the service
	createdJob, err := jobService.CreateJob(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create job",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdJob)
}

// GetAllJobs retrieves a list of worker jobs with optional filtering
func GetAllJobs(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	cfg := c.Locals("config").(*config.Config)

	jobService := job.NewService(db, cfg)

	// Parse query parameters
	options := job.GetAllJobsOptions{
		JobType: c.Query("job_type"),
		Status:  c.Query("status"),
		Limit:   c.QueryInt("limit", 100),
		Offset:  c.QueryInt("offset", 0),
	}

	// Get jobs using the service
	jobs, err := jobService.GetAllJobs(options)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(jobs)
}

// GetJob retrieves a specific job by ID
func GetJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	cfg := c.Locals("config").(*config.Config)

	jobService := job.NewService(db, cfg)
	jobID := c.Params("id")

	job, err := jobService.GetJobByID(jobID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Job not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.JSON(job)
}

// CancelJob cancels a job by changing its status to failed
func CancelJob(c *fiber.Ctx) error {
	db := c.Locals("db").(*gorm.DB)
	cfg := c.Locals("config").(*config.Config)

	jobService := job.NewService(db, cfg)
	jobID := c.Params("id")

	job, err := jobService.CancelJob(jobID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Job not found"})
		}
		if err == gorm.ErrInvalidValue {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Cannot cancel job that is not in pending or retry status",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to cancel job",
			"error":   err.Error(),
		})
	}

	return c.JSON(job)
}
