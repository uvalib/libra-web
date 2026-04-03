package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	type contributor struct {
		FirstName   string
		LastName    string
		Department  string
		Institution string
		ORCID       struct {
			URL string
			ID  string
		}
	}

	var viewData struct {
		WorkID             string
		Visibility         string
		EmbargoReleaseDate string
		Title              string
		Views              int
		Author             contributor
		Advisors           []contributor
		Program            string
		Abstract           string
		Degree             string
		Keywords           string
		Sponsors           []string
		RelatedURLs        []string
		Notes              string
		Language           string
		SignedIn           bool
		ThisYear           string
		Files              []workFile
	}

	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		userInfo := fmt.Sprintf("user %s", jwt.ComputeID)
		log.Printf("INFO: work %s accessed by signed in %s", workID, userInfo)
		viewData.SignedIn = true
	}

	viewData.WorkID = workID
	viewData.Visibility = etdWork.Visibility
	viewData.ThisYear = fmt.Sprintf("%d", time.Now().Year())
	viewData.Title = etdWork.Title
	viewData.Views = etdWork.Views
	viewData.Author.FirstName = etdWork.Author.FirstName
	viewData.Author.LastName = etdWork.Author.LastName
	viewData.Author.Institution = etdWork.Author.Institution
	viewData.Author.ORCID.URL = etdWork.Author.ORCID
	if etdWork.Author.ORCID != "" {
		bits := strings.Split(viewData.Author.ORCID.URL, "/")
		viewData.Author.ORCID.ID = bits[len(bits)-1]
	}
	viewData.Program = etdWork.Program
	for _, adv := range etdWork.Advisors {
		viewData.Advisors = append(viewData.Advisors, contributor{
			LastName:    adv.LastName,
			FirstName:   adv.FirstName,
			Department:  adv.Department,
			Institution: adv.Institution,
		})
	}
	viewData.Abstract = etdWork.Abstract
	viewData.Degree = etdWork.Degree
	if len(etdWork.Keywords) > 0 {
		viewData.Keywords = strings.Join(etdWork.Keywords, "; ")
	}
	viewData.Sponsors = etdWork.Sponsors

	// FIXME? the vue logic has mess regex to see if a url is a url (extractLink) unknown if it is needed
	// It is needed: https://jira.admin.virginia.edu/browse/TDG-2072
	viewData.RelatedURLs = etdWork.RelatedURLs
	viewData.Notes = etdWork.Notes
	viewData.Language = etdWork.Language

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
