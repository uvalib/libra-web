package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
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

func (svc *serviceContext) downloadPublishedFile(c *gin.Context) {
	workID := c.Param("id")
	tgtFile := c.Query("file")
	log.Printf("INFO: request to download file %s from published work %s", tgtFile, workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: get %s work %s for download failed: %s", svc.Namespace, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: check access to download files for work %s", workID)
	access := svc.canAccessWork(c, tgtObj)
	if access.files == false {
		log.Printf("INFO: invalid request to dowload file %s from work %s", tgtFile, workID)
		c.String(http.StatusForbidden, fmt.Sprintf("invalid request to download %s", tgtFile))
		return
	}

	var dlFile uvaeasystore.EasyStoreBlob
	for _, oaFile := range tgtObj.Files() {
		if oaFile.Name() == tgtFile {
			dlFile = oaFile
		}
	}

	if dlFile == nil {
		log.Printf("INFO: file %s not found in %s work %s", tgtFile, svc.Namespace, workID)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", tgtFile))
		return
	}

	// publish download event for public views of works that are not draft
	log.Printf("INFO: download requested from a public view; track it")
	dlDetail := getPublicRequestEventDetails(c, tgtFile)
	evt := uvalibrabus.UvaBusEvent{
		EventName:  uvalibrabus.EventContentDownload,
		Namespace:  tgtObj.Namespace(),
		Identifier: tgtObj.Id(),
		Detail:     dlDetail,
	}
	log.Printf("INFO: publish download event %s", dlDetail)
	if svc.Events.DevMode {
		log.Printf("INFO: dev mode work %s send download event to bus [%s] with source [%s]", workID, svc.Events.BusName, svc.Events.EventSource)
	} else {
		err := svc.Events.Bus.PublishEvent(&evt)
		if err != nil {
			log.Printf("ERROR: unable to publish download event %+v : %s", evt, err.Error())
		}
	}

	// redirect to the newly generated url so the client automatically does the download
	// with no additional JS logic needed
	log.Printf("INFO: %s download link %s", tgtFile, dlFile.Url())
	c.Redirect(http.StatusTemporaryRedirect, dlFile.Url())
}
