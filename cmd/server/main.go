package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"

	// "github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/handlers"
	mngmt "fundermaps/app/handlers/management"
	"fundermaps/app/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	store := session.New(session.Config{
		CookieSecure:   cfg.AuthSecure,
		CookieDomain:   cfg.AuthDomain,
		CookieHTTPOnly: true,
		Expiration:     time.Duration(cfg.AuthExpiration) * time.Hour,
		KeyLookup:      "cookie:session_id",
		CookieSameSite: "Lax",
	})

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: cfg.ProxyEnabled,
		TrustedProxies:          cfg.ProxyNetworks,
		ProxyHeader:             cfg.ProxyHeader,
	})

	app.Use(compress.New())
	app.Use(helmet.New())
	app.Use(recover.New())

	app.Use(healthcheck.New())
	app.Use(favicon.New(favicon.Config{
		File: "./static/favicon.ico",
	}))

	app.Use(middleware.RobotsMiddleware("./static/robots.txt"))

	// app.Use(cors.New(cors.Config{
	// 	AllowHeaders: "Authorization, Content-Type, Accept, X-Requested-With, X-Session-ID",
	// 	MaxAge:       300,
	// }))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", cfg)
		c.Locals("db", db)
		c.Locals("store", store)
		return c.Next()
	})

	app.Use(logger.New(logger.Config{
		Format: "${latency} | ${status} | ${method} | ${path}\n",
	}))

	app.Get("/auth/login", func(c *fiber.Ctx) error {
		return c.SendFile("./public/login.html")
	})
	app.Post("/auth/login", limiter.New(limiter.Config{Max: 50}), handlers.LoginWithForm)
	app.Get("/auth/logout", handlers.Logout)

	// General API
	api := app.Group("/api")
	api.Get("/app/:application_id?", handlers.GetApplication)
	api.Get("/data/contractor", middleware.AuthMiddleware, handlers.GetAllContractors) // TODO: Why not add the contractors to the application data?

	// Auth API
	auth := api.Group("/auth", limiter.New(limiter.Config{Max: 50}))
	auth.Post("/signin", handlers.SigninWithPassword)
	auth.Post("/token-refresh", middleware.AuthMiddleware, handlers.RefreshToken)
	auth.Post("/change-password", middleware.AuthMiddleware, handlers.ChangePassword)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/reset-password", handlers.ResetPassword)

	// OAuth2 API
	oauth2 := api.Group("/v1/oauth2", limiter.New(limiter.Config{Max: 50}))
	oauth2.Get("/authorize", handlers.AuthorizationRequest)
	oauth2.Post("/token", handlers.TokenRequest)
	oauth2.Get("/userinfo", middleware.AuthMiddleware, handlers.GetUserInfo)

	// User API
	user := api.Group("/user", middleware.AuthMiddleware)
	user.Get("/me", handlers.GetCurrentUser)
	user.Put("/me", handlers.UpdateCurrentUser)
	user.Get("/metadata", handlers.GetCurrentUserMetadata)
	user.Put("/metadata", handlers.UpdateCurrentUserMetadata)

	// Mapset API
	mapset := api.Group("/mapset", middleware.AuthMiddleware)
	mapset.Get("/:mapset_id?", handlers.GetMapset)

	// Incident API
	incident := api.Group("/incident", limiter.New(limiter.Config{Max: 50}))
	incident.Post("/", handlers.CreateIncident)
	incident.Post("/upload", handlers.UploadFiles)

	// Geocoder API
	geocoder := api.Group("/geocoder/:geocoder_id", limiter.New(limiter.Config{Max: 50}))
	geocoder.Get("/", handlers.GetGeocoder)
	geocoder.Get("/address", handlers.GetAllAddresses)

	// Product API
	product := api.Group("/product/:building_id", middleware.AuthMiddleware) // TODO: requestid.New() use when serving the public API
	product.Get("/analysis", middleware.TrackerMiddleware, handlers.GetAnalysis)
	product.Get("/statistics", handlers.GetStatistics)
	product.Get("/subsidence", handlers.GetDataSubsidence) // TODO: There may be no need for this endpoint
	product.Get("/subsidence/historic", handlers.GetDataSubsidenceHistoric)

	// Report API
	report := api.Group("/report/:building_id", middleware.AuthMiddleware)
	report.Get("/", handlers.GetReport)

	// Inquiry API
	inquiry := api.Group("/inquiry", middleware.AuthMiddleware)
	inquiry.Post("/", handlers.CreateInquiry)
	inquiry.Post("/:inquiry_id", handlers.CreateInquirySample)

	// Recovery API
	recovery := api.Group("/recovery", middleware.AuthMiddleware)
	recovery.Post("/", handlers.CreateRecovery)
	recovery.Post("/:recovery_id", handlers.CreateRecoverySample)

	// PDF API
	pdf := api.Group("/pdf", middleware.AuthMiddleware)
	pdf.Get("/:id", handlers.GetPDF)

	// TODO: Drop the 'v1' from the URL
	// Management API
	management := api.Group("/v1/management", middleware.AuthMiddleware, middleware.AdminMiddleware)
	management.Get("/app", mngmt.GetAllApplications)
	management.Post("/app", mngmt.CreateApplication)
	management_app := management.Group("/app/:app_id")
	management_app.Get("/", mngmt.GetApplication)
	management_app.Put("/", mngmt.UpdateApplication)
	management.Get("/mapset", mngmt.GetAllMapsets)
	management_mapset := management.Group("/mapset/:mapset_id")
	management_mapset.Get("/", mngmt.GetMapsetByID)
	management_incident := management.Group("/incident/:incident_id")
	management_incident.Delete("/", mngmt.DeleteIncident)
	management.Get("/user", mngmt.GetAllUsers)
	management.Post("/user", mngmt.CreateUser)
	// management.Get("/user/:email", handlers.GetUserByEmail)
	management_user := management.Group("/user/:user_id")
	management_user.Get("/", mngmt.GetUser)
	management_user.Put("/", mngmt.UpdateUser)
	management_user.Get("/api-key", mngmt.CreateApiKey)
	management_user.Post("/reset-password", mngmt.ResetUserPassword)

	// Job management routes
	management.Get("/jobs", mngmt.GetAllJobs)
	management.Post("/jobs", mngmt.CreateJob)
	management_job := management.Group("/jobs/:id")
	management_job.Get("/", mngmt.GetJob)
	management_job.Post("/cancel", mngmt.CancelJob)
	management.Post("/org", mngmt.CreateOrganization)
	management.Get("/org", mngmt.GetAllOrganizations)
	// management.Get("/org/:name", mngmt.GetOrganizationByName) # TODO: Implement
	management_org := management.Group("/org/:org_id")
	management_org.Get("/", mngmt.GetOrganization)
	management_org.Put("/", mngmt.UpdateOrganization)
	management_org_mapset := management_org.Group("/mapset")
	// management_org_mapset.Get("/", mngmt.GetAllOrganizationMapsets)
	management_org_mapset.Post("/", mngmt.AddMapsetToOrganization)
	management_org_mapset.Delete("/", mngmt.RemoveMapsetFromOrganization)
	management_org_user := management_org.Group("/user")
	management_org_user.Get("/", mngmt.GetAllOrganizationUsers)
	management_org_user.Post("/", mngmt.AddUserToOrganization)
	management_org_user.Delete("/", mngmt.RemoveUserFromOrganization)

	// Diagnostic API
	diag := api.Group("/diag")
	diag.Get("/ip", handlers.GetIP)
	diag.Get("/req", handlers.GetHeaders)

	app.Use(limiter.New(limiter.Config{}), func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not found"})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.ServerPort)))
}
