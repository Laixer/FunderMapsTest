package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"fundermaps/app/config"
	"fundermaps/app/database"
	"fundermaps/app/handlers"
	mngmt "fundermaps/app/handlers/management"
	"fundermaps/app/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: All of this proxy stuff should be configurable
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"10.0.0.0/8", "fc00::/7"},
		ProxyHeader:             "Do-Connecting-Ip",
	})

	app.Use(compress.New())
	app.Use(helmet.New())
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

	app.Use(logger.New(logger.Config{
		Format: "${method} | ${status} | ${latency} | ${ip} | ${path}\n",
	}))

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

	// Mapset application
	mapset := api.Group("/mapset", limiter.New(limiter.Config{Max: 50}))
	mapset.Get("/:mapset_id?", middleware.AuthMiddleware, handlers.GetMapset) // TODO: Does not work for public mapsets

	// Incident application
	incident := api.Group("/incident")
	incident.Post("/", handlers.CreateIncident)
	incident.Post("/upload", handlers.UploadFiles)

	// Management API
	management := api.Group("/v1/management", middleware.AuthMiddleware, middleware.AdminMiddleware)
	management.Get("/app", mngmt.GetAllApplications)
	management.Post("/app", mngmt.CreateApplication)
	management.Get("/mapset", mngmt.GetAllMapsets)
	management_mapset := management.Group("/mapset/:mapset_id")
	management_mapset.Get("/", mngmt.GetMapsetByID)
	management.Get("/user", mngmt.GetAllUsers)
	management.Post("/user", mngmt.CreateUser)
	// management.Get("/user/:email", handlers.GetUserByEmail)
	management_user := management.Group("/user/:user_id")
	management_user.Get("/", mngmt.GetUser)
	management_user.Put("/", mngmt.UpdateUser)
	management_user.Get("/api-key", mngmt.CreateApiKey)
	management_user.Post("/reset-password", mngmt.ResetUserPassword)
	management.Post("/org", handlers.CreateOrganization)
	management.Get("/org", handlers.GetAllOrganizations)
	// management.Get("/org/:name", handlers.GetOrganizationByName) # TODO: Implement
	management_org := management.Group("/org/:org_id")
	management_org.Get("/", handlers.GetOrganization)
	management_org_mapset := management_org.Group("/mapset")
	// management_org_mapset.Get("/", handlers.GetAllOrganizationMapsets)
	management_org_mapset.Post("/", handlers.AddMapsetToOrganization)
	management_org_mapset.Delete("/", handlers.RemoveMapsetFromOrganization)
	management_org_user := management_org.Group("/user")
	management_org_user.Get("/", handlers.GetAllOrganizationUsers)
	management_org_user.Post("/", handlers.AddUserToOrganization)
	management_org_user.Delete("/", handlers.RemoveUserFromOrganization)

	geocoder := api.Group("/geocoder/:geocoder_id", limiter.New(limiter.Config{Max: 50}))
	geocoder.Get("/", handlers.GetGeocoder)
	geocoder.Get("/address", handlers.GetAllAddresses)

	// TODO: Add tracker middleware
	product := api.Group("/v4/product/:building_id", middleware.AuthMiddleware, requestid.New())
	product.Get("/analysis", middleware.TrackerMiddleware, handlers.GetAnalysis)
	// product.Get("/statistics", handlers.GetAnalysis)
	// product.Get("/subsidence", handlers.GetDataSubsidence)
	product.Get("/subsidence/historic", handlers.GetDataSubsidenceHistoric)

	diag := api.Group("/diag")
	diag.Get("/ip", handlers.GetIP)
	diag.Get("/req", handlers.GetHeaders)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not found"})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.ServerPort)))
}
