package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	type relatedURL struct {
		IsURL bool
		Value string
	}

	var viewData struct {
		BaseURL            string
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
		RelatedURLs        []relatedURL
		Notes              string
		Language           string
		Citation           string
		PublishedDate      string
		PersistentLink     string
		ThisYear           string
		Files              []workFile
		License            struct {
			Name string
			URL  string
		}
	}

	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		userInfo := fmt.Sprintf("user %s", jwt.ComputeID)
		log.Printf("INFO: public view of work %s accessed by signed in user %s", workID, userInfo)
	} else {
		log.Printf("INFO: public view of work %s accessed by anonymous user", workID)
	}

	viewData.BaseURL = svc.EtdURL
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
	for _, relVal := range etdWork.RelatedURLs {
		entry := relatedURL{
			IsURL: true,
			Value: strings.TrimSpace(relVal),
		}
		if _, err := url.ParseRequestURI(entry.Value); err != nil {
			entry.IsURL = false
		}
		viewData.RelatedURLs = append(viewData.RelatedURLs, entry)
	}
	viewData.Notes = etdWork.Notes
	viewData.Language = etdWork.Language
	viewData.License.Name = etdWork.License
	viewData.License.URL = etdWork.LicenseURL
	viewData.PersistentLink = etdWork.PersistentLink

	pubDate, dErr := time.Parse(svc.TimeFormat, etdWork.PublishedAt)
	if dErr != nil {
		log.Printf("ERROR: unable to parse publised date: %s", dErr.Error())
	} else {
		viewData.PublishedDate = pubDate.Format("2006-01-02")
	}

	if etdWork.Visibility == "uva" || etdWork.Visibility == "embargo" {
		endDate, _ := time.Parse(svc.TimeFormat, etdWork.Embargo.ReleaseDate)
		viewData.EmbargoReleaseDate = endDate.Format("2006-01-02")
	}

	// build suggested citaion
	//[Author LastName], [Author FirstName]. [Title]. [Author Institution], [program], [Degree], [Published Year], [DOI URI].
	viewData.Citation = fmt.Sprintf("%s, %s. %s. %s, %s, %s, %s, %s.",
		etdWork.Author.LastName, etdWork.Author.FirstName,
		etdWork.Title, etdWork.Author.Institution, etdWork.Program, etdWork.Degree,
		viewData.PublishedDate, viewData.PersistentLink)

	// add files to the work data
	for _, f := range etdWork.Files {
		viewData.Files = append(viewData.Files, workFile{FileName: f.Name, Downloads: f.Downloads})
	}

	// render template as html using the data set up above
	c.HTML(http.StatusOK, "view.html", viewData)
}
