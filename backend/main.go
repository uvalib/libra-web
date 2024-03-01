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
const Version = "0.0.1"

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
	router.GET("/config", svc.getConfig)
	router.GET("/healthcheck", svc.healthCheck)
	router.GET("/version", svc.getVersion)

	api := router.Group("/api", svc.authMiddleware)
	{
		api.GET("/lookup/:cid", svc.lookupComputeID)

		// NOTE: when a deposit is requested, a temporary work token is generated
		// this token is used as a subrirectory to stage uploaded files. Upon submissions,
		// the files from the token directory are added to the work and the token is converted to a libra ID
		api.GET("/token", svc.getDepositToken)
		api.POST("/upload/:token", svc.uploadSubmissionFiles)
		api.DELETE("/:token/:filename", svc.removeSubmissionFile)
		api.POST("/cancel/:token", svc.cancelSubmission)
		api.POST("/oa/:token", svc.oaSubmit)
		api.POST("/etd/:token", svc.etdSubmit)

		// After initial submission, the work is referenced by the permanent ID
		api.PUT("/oa/:id", svc.oaUpdate)
		api.DELETE("/works/:id", svc.deleteWork)
	}

	// Note: in dev mode, this is never actually used. The front end is served
	// by node/vite and it proxies all requests to the API to the routes above
	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	// // add a catchall route that renders the index page.
	// router.NoRoute(func(c *gin.Context) {
	// 	c.File("./public/index.html")
	// })

	portStr := fmt.Sprintf(":%d", cfg.port)
	versionMap := svc.lookupVersion()
	versionStr := fmt.Sprintf("%s-%s", versionMap["version"], versionMap["build"])
	log.Printf("INFO: start Libra3 v%s on port %s with CORS support enabled", versionStr, portStr)
	log.Fatal(router.Run(portStr))
}
