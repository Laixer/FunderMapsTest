package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"

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

	// Auth API
	api := app.Group("/api")
	auth := api.Group("/auth", limiter.New(limiter.Config{Max: 50}))
	auth.Post("/signin", handlers.SigninWithPassword)
	auth.Post("/token-refresh", middleware.AuthMiddleware, handlers.RefreshToken)

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

	// User management routes
	management.Get("/user", mngmt.GetAllUsers)
	management.Post("/user", mngmt.CreateUser)
	management_user := management.Group("/user/:user_id")
	management_user.Get("/", mngmt.GetUser)
	management_user.Put("/", mngmt.UpdateUser)
	management_user.Delete("/", mngmt.DeleteUser)
	management_user.Get("/api-key", mngmt.GetApiKeys)
	management_user.Post("/api-key", mngmt.CreateApiKey)
	management_user.Delete("/api-key", mngmt.DeleteApiKey)
	management_user.Post("/reset-password", mngmt.ResetUserPassword)

	// Job management routes
	management.Get("/jobs", mngmt.GetAllJobs)
	management.Post("/jobs", mngmt.CreateJob)
	management_job := management.Group("/jobs/:id")
	management_job.Get("/", mngmt.GetJob)
	management_job.Post("/cancel", mngmt.CancelJob)

	// Organization management routes
	management.Post("/org", mngmt.CreateOrganization)
	management.Get("/org", mngmt.GetAllOrganizations)
	management_org := management.Group("/org/:org_id")
	management_org.Get("/", mngmt.GetOrganization)
	management_org.Put("/", mngmt.UpdateOrganization)
	management_org.Delete("/", mngmt.DeleteOrganization)
	management_org_mapset := management_org.Group("/mapset")
	management_org_mapset.Get("/", mngmt.GetAllOrganizationMapsets)
	management_org_mapset.Post("/", mngmt.AddMapsetToOrganization)
	management_org_mapset.Delete("/", mngmt.RemoveMapsetFromOrganization)
	management_org_user := management_org.Group("/user")
	management_org_user.Get("/", mngmt.GetAllOrganizationUsers)
	management_org_user.Post("/", mngmt.AddUserToOrganization)
	management_org_user.Delete("/", mngmt.RemoveUserFromOrganization)

	// Organization geolock management routes
	management_org_district := management_org.Group("/district")
	management_org_district.Get("/", mngmt.GetOrganizationGeolockDistricts)
	management_org_district.Post("/", mngmt.AddDistrictToOrganization)
	management_org_district.Delete("/", mngmt.RemoveDistrictFromOrganization)
	management_org_municipality := management_org.Group("/municipality")
	management_org_municipality.Get("/", mngmt.GetOrganizationGeolockMunicipalities)
	management_org_municipality.Post("/", mngmt.AddMunicipalityToOrganization)
	management_org_municipality.Delete("/", mngmt.RemoveMunicipalityFromOrganization)
	management_org_neighborhood := management_org.Group("/neighborhood")
	management_org_neighborhood.Get("/", mngmt.GetOrganizationGeolockNeighborhoods)
	management_org_neighborhood.Post("/", mngmt.AddNeighborhoodToOrganization)
	management_org_neighborhood.Delete("/", mngmt.RemoveNeighborhoodFromOrganization)

	app.Use(limiter.New(limiter.Config{}), func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not found"})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.ServerPort)))
}
