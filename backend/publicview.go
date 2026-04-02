package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) getStaticPage(c *gin.Context) {
	id := c.Param("id")
	log.Printf("INFO: generate static pubic view  %s", id)

	var viewData struct {
		WorkID      string
		DisplayName string
		SignedIn    bool
	}

	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		userInfo := fmt.Sprintf("user %s", jwt.ComputeID)
		log.Printf("INFO: work %s accessed by signed in %s", id, userInfo)
		viewData.SignedIn = true
		viewData.DisplayName = jwt.DisplayName
	}

	viewData.WorkID = id
	var rendered bytes.Buffer
	if err := svc.PublicTemplate.Execute(&rendered, viewData); err != nil {
		log.Printf("ERROR: unable to generate public view for %s: %s", id, err.Error())
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, rendered.String())
}
