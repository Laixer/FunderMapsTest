package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		// TrustedProxies:          []string{"10.0.0.0/8"},
		TrustedProxies: []string{"10.244.4.113"},
	})

	app.Use(compress.New())
	app.Use(helmet.New()) // TODO: We only need this for internal routes
	app.Use(recover.New())

	app.Use(healthcheck.New())
	app.Use(favicon.New(favicon.Config{
		File: "./static/favicon.ico",
	}))

	// TODO: Maybe move into middleware
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
		return c.SendFile("./static/robots.txt")
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", cfg)
		c.Locals("db", db)
		return c.Next()
	})

	// TODO: We might want different loggers for different routes
	app.Use(logger.New(logger.Config{
		Format: "${method} | ${status} | ${latency} | ${ip} | ${path}\n",
	}))

	api := app.Group("/api", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=3600")
		return c.Next()
	})
	api.Get("/app/:application_id", handlers.GetApplication) // TODO: Make the parameter optional

	// TODO: Add the limiter middleware
	auth := api.Group("/auth")
	auth.Post("/signin", handlers.SigninWithPassword)
	auth.Get("/token-refresh", middleware.AuthMiddleware, handlers.RefreshToken)
	auth.Post("/change-password", middleware.AuthMiddleware, handlers.ChangePassword)
	// auth.Post("/forgot-password", handlers.ForgotPassword)
	// auth.Post("/reset-password", handlers.ResetPassword)

	user := api.Group("/user", middleware.AuthMiddleware)
	user.Get("/me", handlers.GetCurrentUser) // Return User + Organization + Organization Role
	user.Put("/me", handlers.UpdateCurrentUser)
	user.Get("/metadata", handlers.GetCurrentUserMetadata)
	// user.Put("/metadata", handlers.UpdateCurrentUserMetadata)

	management := api.Group("/v1/management", middleware.AuthMiddleware) // middleware.AdminMiddleware
	management.Post("/create-app", handlers.CreateApplication)
	management.Post("/create-user", handlers.CreateUser)
	management.Post("/create-org", handlers.CreateOrganization)
	management.Post("/create-auth-token", handlers.CreateAuthKey)
	management.Post("/add-user-to-org", handlers.AddUserToOrganization)
	// management.Post("/remove-user-from-org", handlers.RemoveUserFromOrganization)
	management.Post("/add-mapset-to-org", handlers.AddMapsetToOrganization)

	geocoder := api.Group("/geocoder", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, max-age=86400")
		return c.Next()
	})
	geocoder.Get("/:geocoder_id", handlers.GetGeocoder)

	// TODO: Needs 'user,admin' role
	// api.Get("/incident", middleware.AuthMiddleware, handlers.GetIncident)
	api.Post("/incident", handlers.CreateIncident)
	api.Get("/contractor", middleware.AuthMiddleware, handlers.GetAllContractors)
	api.Get("/mapset/:mapset_id?", middleware.AuthMiddleware, handlers.GetMapset)

	product := api.Group("/v4/product", middleware.AuthMiddleware, requestid.New())
	product.Get("/analysis/:building_id", handlers.GetAnalysis)
	product.Get("/statistics/:building_id", handlers.GetAnalysis)

	test := api.Group("/test")
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.ServerPort)))
}
