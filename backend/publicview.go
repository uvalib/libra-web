package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) getStaticPage(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: generate static pubic view  %s", workID)

	etdWork, err := svc.getWork(c, workID, "view")
	if err != nil {
		log.Printf("ERROR: get work %s failed: %d - %s", workID, err.StatusCode, err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}

	var viewData struct {
		WorkID      string
		Title       string
		Views       int
		DisplayName string
		SignedIn    bool
		ThisYear    string
	}

	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		userInfo := fmt.Sprintf("user %s", jwt.ComputeID)
		log.Printf("INFO: work %s accessed by signed in %s", workID, userInfo)
		viewData.SignedIn = true
		viewData.DisplayName = jwt.DisplayName
	}

	viewData.WorkID = workID
	viewData.ThisYear = fmt.Sprintf("%d", time.Now().Year())
	viewData.Title = etdWork.Title
	viewData.Views = etdWork.Views
	var rendered bytes.Buffer
	if err := svc.ViewTemplates.Execute(&rendered, viewData); err != nil {
		log.Printf("ERROR: unable to generate public view for %s: %s", workID, err.Error())
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, rendered.String())
}
