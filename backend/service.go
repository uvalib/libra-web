package main

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// serviceContext contains common data used by all handlers
type serviceContext struct {
	Version     string
	JWTKey      string
	DevAuthUser string
}

// InitializeService sets up the service context for all API handlers
func initializeService(version string, cfg *configData) *serviceContext {
	ctx := serviceContext{Version: version,
		JWTKey:      cfg.jwtKey,
		DevAuthUser: cfg.devAuthUser}

	return &ctx
}

// GetVersion reports the version of the serivce
func (svc *serviceContext) getVersion(c *gin.Context) {
	vMap := svc.lookupVersion()
	c.JSON(http.StatusOK, vMap)
}

func (svc *serviceContext) lookupVersion() map[string]string {
	build := "unknown"
	// working directory is the bin directory, and build tag is in the root
	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = svc.Version
	vMap["build"] = build
	return vMap
}

// HealthCheck reports the health of the serivce
func (svc *serviceContext) healthCheck(c *gin.Context) {
	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}
	hcMap := make(map[string]hcResp)
	hcMap["libra3"] = hcResp{Healthy: true}

	c.JSON(http.StatusOK, hcMap)
}
