package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Version of the service
const Version = "0.9.0"

func main() {
	// Load cfg
	log.Printf("===> Libra3 is starting up <===")
	cfg := getConfiguration()
	svc := initializeService(Version, cfg)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Set routes and start serve
	router.GET("/authenticate", svc.authenticate)
	router.GET("/authcheck", svc.checkAuthToken)
	router.GET("/config", svc.getConfig)
	router.GET("/healthcheck", svc.healthCheck)
	router.GET("/version", svc.getVersion)
	router.GET("/sitemap.xml", svc.GetSitemap)

	api := router.Group("/api", svc.userMiddleware)
	{
		api.GET("/users/lookup/:cid", svc.lookupComputeID)
		api.GET("/users/orcid/:cid", svc.lookupOrcidID)

		// NOTE: when a deposit is requested, a temporary work token is generated
		// this token is used as a subdirectory to stage uploaded files. Upon submissions,
		// the files from the token directory are added to the work and the token is converted to a libra ID
		api.GET("/token", svc.getDepositToken)
		api.POST("/upload/:token", svc.uploadSubmissionFiles)
		api.DELETE("/:token/:filename", svc.removeSubmissionFile)
		api.POST("/cancel/:token", svc.cancelSubmission)

		api.GET("/audits/:id", svc.getAudits)

		// After initial submission, the work is referenced by the permanent ID
		api.GET("/works/:id", svc.getWork)
		api.GET("/works/:id/files/:name", svc.downloadFile)
		api.PUT("/works/:id", svc.updateWork)
		api.POST("/works/:id/publish", svc.publishWork)

		// user search of all works
		api.GET("/works/search", svc.userSearch)

		// register users for optional works. only available for admin or registrar users
		api.POST("/register", svc.registrarMiddleware, svc.submitOptionalRegistrations)

		admin := api.Group("/admin", svc.adminMiddleware)
		{
			admin.POST("/impersonate/:computeID", svc.adminImpersonateUser)
			admin.GET("/search", svc.adminSearch)
			admin.DELETE("/works/:id", svc.adminDeleteWork)
			admin.DELETE("/works/:id/publish", svc.adminUnpublishWork)
		}
	}

	// Note: in dev mode, this is never actually used. The front end is served
	// by node/vite and it proxies all requests to the API to the routes above
	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	// add a catchall route that renders the index page.
	router.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	portStr := fmt.Sprintf(":%d", cfg.port)
	versionMap := svc.lookupVersion()
	versionStr := fmt.Sprintf("%s-%s", versionMap["version"], versionMap["build"])
	log.Printf("INFO: start Libra3 v%s on port %s with CORS support enabled", versionStr, portStr)
	log.Fatal(router.Run(portStr))
}
