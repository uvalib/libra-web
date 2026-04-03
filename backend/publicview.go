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

	type workFile struct {
		FileName  string
		Downloads int
	}

	var viewData struct {
		WorkID             string
		Visibility         string
		EmbargoReleaseDate string
		Title              string
		Views              int
		DisplayName        string
		SignedIn           bool
		ThisYear           string
		Files              []workFile
	}

	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		userInfo := fmt.Sprintf("user %s", jwt.ComputeID)
		log.Printf("INFO: work %s accessed by signed in %s", workID, userInfo)
		viewData.SignedIn = true
		viewData.DisplayName = jwt.DisplayName
	}

	viewData.WorkID = workID
	viewData.Visibility = etdWork.Visibility
	viewData.ThisYear = fmt.Sprintf("%d", time.Now().Year())
	viewData.Title = etdWork.Title
	viewData.Views = etdWork.Views

	if etdWork.Visibility == "uva" || etdWork.Visibility == "embargo" {
		endDate, _ := time.Parse(svc.TimeFormat, etdWork.Embargo.ReleaseDate)
		viewData.EmbargoReleaseDate = endDate.Format("2006-01-02")
	}

	for _, f := range etdWork.Files {
		viewData.Files = append(viewData.Files, workFile{FileName: f.Name, Downloads: f.Downloads})
	}

	var rendered bytes.Buffer
	if err := svc.ViewTemplates.Execute(&rendered, viewData); err != nil {
		log.Printf("ERROR: unable to generate public view for %s: %s", workID, err.Error())
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, rendered.String())
}
