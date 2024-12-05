package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"fundermaps/internal/config"
	"fundermaps/internal/database"
	"fundermaps/internal/handlers"
	"fundermaps/internal/mail"
	"fundermaps/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(compress.New())
	app.Use(helmet.New())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(healthcheck.New())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", cfg)
		c.Locals("db", db)
		return c.Next()
	})

	api := app.Group("/api")
	api.Get("/app/:application_id", handlers.GetApplication)

	auth := api.Group("/auth")
	auth.Post("/signin", handlers.SigninWithPassword)
	// auth.Post("/hash", handlers.Hash)
	auth.Get("/token-refresh", middleware.AuthMiddleware, handlers.RefreshToken)

	user := api.Group("/user", middleware.AuthMiddleware)
	user.Get("/me", handlers.GetCurrentUser) // Return User + Organization + Organization Role
	user.Put("/me", handlers.UpdateUser)

	admin := api.Group("/admin", middleware.AuthMiddleware) // middleware.AdminMiddleware
	admin.Post("/create-user", handlers.CreateUser)
	admin.Post("/create-org", handlers.CreateOrganization)
	admin.Post("/create-auth-token", handlers.CreateAuthKey)

	// geocoder := api.Group("/geocoder")
	// geocoder.Get("/address/:address", handlers.GetAddress) // TODO: Maybe obsolete
	// geocoder.Get("/building/:building", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/residence/:residence", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/neighborhood/:neighborhood", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/district/:district", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/municipality/:municipality", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/state/:state", handlers.GetBuilding) // TODO: Maybe obsolete
	// geocoder.Get("/:id", handlers.GetGeocoder)

	// incident := api.Group("incident")
	api.Post("/incident", handlers.CreateIncident)

	// TODO: Add another middleware to check if user is role 'service' or 'admin'
	product := api.Group("v4/product", middleware.AuthMiddleware)
	product.Get("/analysis/:building_id", handlers.GetAnalysis)
	product.Get("/statistics/:building_id", handlers.GetAnalysis)

	test := api.Group("/test")
	// test.Get("/db-test", handlers.GetAllContractors)
	// test.Get("/:short_code", handlers.GetRewriteUrl)
	test.Post("/mail", func(c *fiber.Ctx) error {
		type EmailInput struct {
			Subject string `json:"subject"`
			Body    string `json:"body"`
			From    string `json:"from"`
			To      string `json:"to"`
		}

		var input EmailInput
		if err := c.BodyParser(&input); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if input.Subject == "" || input.Body == "" || input.From == "" || input.To == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		message := mail.Email{
			Subject: input.Subject,
			Body:    input.Body,
			From:    input.From,
			To:      []string{input.To},
		}

		mailer := mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase)
		mailer.SendMail(&message)

		return c.SendStatus(fiber.StatusOK)
	})

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.ServerPort)))
}
