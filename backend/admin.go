package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

func (svc *serviceContext) returnCSV(c *gin.Context, jsonResp *indexResp) {
	c.Header("Content-Type", "text/csv")
	cw := csv.NewWriter(c.Writer)
	csvHead := []string{
		"Id", "Program", "Degree", "Title", "Depositor",
		"AuthorID", "Author First Name", "Author Last Name", "Author Institution", "Author ORCiD",
		"Advisors", "Abstract", "Rights", "Keywords", "Language", "Related Links", "Sponsoring Agency",
		"Notes", "Admin Notes",
		"Create Date", "Last Modified Date", "Published Date", "Embargo State", "Embargo End Date",
		"DOI", "Source", "Views", "Downloads"}
	cw.Write(csvHead)
	for _, work := range jsonResp.Hits {
		currVis := svc.calculateVisibility(work.Fields.DefaultVisibility, work.Fields.EmbargoRelaeseDate, work.Fields.EmbargoRelaeseVisibility)
		advisors := make([]string, 0)
		for i, adv := range work.Metadata.Advisors {
			adv := fmt.Sprintf("%d: (%s) %s %s %s %s", i, adv.ComputeID, adv.FirstName, adv.LastName, adv.Department, adv.Institution)
			advisors = append(advisors, adv)
		}

		orcidInfo, oErr := svc.doOrcidLookup(work.Metadata.Author.ComputeID)
		if oErr != nil {
			log.Printf("ERROR: unable to obtain orcid info for user %s: %s", work.Metadata.Author.ComputeID, oErr.Error())
		}
		metrics, mErr := svc.getPublicViewMetrics(work.ID)
		if mErr != nil {
			log.Printf("ERROR: unable to get view metrics %s", mErr.Error())
		}

		line := make([]string, 0)
		line = append(line, work.ID)
		line = append(line, work.Metadata.Program)
		line = append(line, work.Metadata.Degree)
		line = append(line, work.Metadata.Title)
		line = append(line, work.Fields.Depositor)
		line = append(line, work.Metadata.Author.ComputeID)
		line = append(line, work.Metadata.Author.FirstName)
		line = append(line, work.Metadata.Author.LastName)
		line = append(line, work.Metadata.Author.Institution)
		if orcidInfo != nil {
			line = append(line, orcidInfo.Orcid)
		} else {
			line = append(line, "")
		}
		line = append(line, strings.Join(advisors, "\n"))
		line = append(line, work.Metadata.Abstract)
		line = append(line, work.Metadata.License)
		line = append(line, strings.Join(work.Metadata.Keywords, "; "))
		line = append(line, work.Metadata.Language)
		line = append(line, strings.Join(work.Metadata.RelatedURLs, "; "))
		line = append(line, strings.Join(work.Metadata.Sponsors, "; "))
		line = append(line, work.Metadata.Notes)
		line = append(line, work.Metadata.AdminNotes)
		line = append(line, work.Fields.CreateDate)
		line = append(line, work.Fields.ModifyDate)
		line = append(line, work.Fields.PublishDate)
		line = append(line, currVis)
		line = append(line, work.Fields.EmbargoRelaeseDate)
		line = append(line, work.Fields.Doi)
		line = append(line, work.Fields.SourceID)
		if metrics != nil {
			line = append(line, fmt.Sprintf("%d", metrics.Views))
			dl := 0
			for _, f := range metrics.Files {
				dl += f.Downloads
			}
			line = append(line, fmt.Sprintf("%d", dl))
		} else {
			// no metrics data found, so zero for view and download
			line = append(line, "0")
			line = append(line, "0")
		}
		cw.Write(line)
	}

	cw.Flush()
}

func (svc *serviceContext) adminSearch(c *gin.Context) {
	recentActivity := (c.Query("recent") != "")
	qStr := c.Query("q")
	if qStr == "" || recentActivity {
		qStr = "*"
	}

	offset, pageErr := strconv.ParseInt(c.Query("offset"), 10, 0)
	if pageErr != nil {
		log.Printf("ERROR: invalid offset parameter %s: %s", c.Query("offset"), pageErr.Error())
		return
	}
	limit, pageErr := strconv.ParseInt(c.Query("limit"), 10, 0)
	if pageErr != nil {
		log.Printf("ERROR: invalid  limit parameter %s: %s", c.Query("offset"), pageErr.Error())
		return
	}

	exportCount, _ := strconv.ParseInt(c.Query("export"), 10, 0)
	if exportCount > 0 {
		log.Printf("INFO: export of %d records requested", exportCount)
		limit = exportCount
	}

	filters := make([]string, 0)
	payload := map[string]any{"offset": offset, "limit": limit}
	if recentActivity {
		log.Printf("INFO: search for works modified in the last 7 days")
		payload["q"] = "*"
		sevenDaysAgoUnix := time.Now().AddDate(0, 0, -7).Unix()
		filters = append(filters, fmt.Sprintf("modifiedUnix >= %d", sevenDaysAgoUnix))
		payload["sort"] = []string{"modified:desc"}
	} else {
		log.Printf("INFO: admin search for works with [%s]", qStr)
		payload["q"] = qStr
		if c.Query("sort") != "" {
			sort := c.Query("sort")
			switch sort {
			case "created":
				sort = "fields.create-date"
			case "title":
				sort = "metadata.title"
			case "published":
				sort = "fields.publish-date"
			}
			payload["sort"] = []string{fmt.Sprintf("%s:%s", sort, c.Query("order"))}
		}

		//Filter Example: "filter": ["fields.draft=true","fields.source=sis"]}
		if c.Query("source") != "" {
			filters = append(filters, fmt.Sprintf("fields.source=%s", c.Query("source")))
		}
		if c.Query("draft") != "" {
			filters = append(filters, fmt.Sprintf("fields.draft=%s", c.Query("draft")))
		}

		pubDateQ := getUnixDateQuery("publishedUnix", c.Query("published"))
		if pubDateQ != "" {
			filters = append(filters, pubDateQ)
		}
		createDateQ := getUnixDateQuery("createdUnix", c.Query("created"))
		if createDateQ != "" {
			filters = append(filters, createDateQ)
		}
	}

	if len(filters) > 0 {
		payload["filter"] = strings.Join(filters, " AND ")
	}

	log.Printf("INFO: search payload [%+v]", payload)
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: search for [%s] failed: %s", qStr, respErr.Message)
		c.String(respErr.StatusCode, respErr.Message)
		return
	}

	var jsonResp indexResp
	err := json.Unmarshal(rawResp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if exportCount > 0 {
		svc.returnCSV(c, &jsonResp)
	} else {
		resp := svc.parseIndexSearchHits(jsonResp)

		log.Printf("INFO: received %d search hits. Elapsed Time: %d (ms)", len(resp.Hits), jsonResp.ProcessingTime)

		c.JSON(http.StatusOK, resp)
	}
}

func getUnixDateQuery(dateField, dateFilter string) string {
	dateQ := ""
	if dateFilter != "" {
		// query format: 2026-01-01 to 2026-12-31
		dateParts := strings.Split(dateFilter, " to ")
		log.Printf("INFO: search for %s from %s to %s", dateField, dateParts[0], dateParts[1])
		parsedDate, _ := time.Parse("2006-01-02", dateParts[0])
		dateQ = fmt.Sprintf("%s >= %d", dateField, parsedDate.Unix())
		parsedDate, _ = time.Parse("2006-01-02", dateParts[1])
		parsedDate = parsedDate.AddDate(0, 0, 1)
		dateQ += fmt.Sprintf(" AND %s <= %d", dateField, parsedDate.Unix())
	}
	return dateQ
}

func (svc *serviceContext) adminImpersonateUser(c *gin.Context) {
	tgtComputeID := c.Param("computeID")
	adminClaims := getJWTClaims(c)
	log.Printf("INFO: admin user %s request to impersonate user %s", adminClaims.ComputeID, tgtComputeID)

	if err := svc.Protected.refreshJWT(svc.JWTKey); err != nil {
		log.Printf("ERROR: unable to refresh protected service jwt for admin impersonate request: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := fmt.Sprintf("%s/user/%s?auth=%s", svc.Protected.UserServiceURL, tgtComputeID, svc.Protected.JWT)
	resp, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("ERROR: unable get info for user %s impersonate request: %s", tgtComputeID, userErr.Message)
		c.String(userErr.StatusCode, userErr.Message)
		return
	}
	var jsonResp userServiceResp
	if err := json.Unmarshal(resp, &jsonResp); err != nil {
		log.Printf("ERROR: unable to parse user serice response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	jsonResp.User.Role = "user"
	signedStr, jwtErr := svc.mintUserJWT(&jsonResp.User)
	if jwtErr != nil {
		log.Printf("ERROR: unable to generate JWT for impersonated user %s: %s", tgtComputeID, jwtErr.Error())
		c.String(http.StatusInternalServerError, jwtErr.Error())
		return
	}

	log.Printf("INFO: impersonate jwt: %s", signedStr)
	c.SetCookie("libra3_impersonate_jwt", signedStr, 10, "/", "", false, false)
	c.SetSameSite(http.SameSiteLaxMode)
	c.String(http.StatusOK, "impersonated")
}

func (svc *serviceContext) adminUpdatePublishedDate(c *gin.Context) {
	workID := c.Param("id")
	var dateReq struct {
		NewDate string `json:"newDate"`
	}
	err := c.ShouldBindJSON(&dateReq)
	if err != nil {
		log.Printf("ERROR: bad payload in published date update request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("INFO: get work %s %s to update published date to %s", svc.Namespace, workID, dateReq.NewDate)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.BaseComponent|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: unable to get work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	claims := getJWTClaims(c)
	svc.auditDatePublished(claims.ComputeID, tgtObj, dateReq.NewDate)

	fields := tgtObj.Fields()
	fields["publish-date"] = dateReq.NewDate
	_, err = svc.EasyStore.ObjectUpdate(tgtObj, uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: update date published for work %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("publish date update failed: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, fields["publish-date"])
}

func (svc *serviceContext) adminUnpublishWork(c *gin.Context) {
	workID := c.Param("id")
	claims := getJWTClaims(c)
	log.Printf("INFO: admin %s requests unpublish work %s", claims.ComputeID, workID)

	log.Printf("INFO: get work %s for unpublish", workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.BaseComponent|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: unable to get work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fields := tgtObj.Fields()
	if fields["draft"] == "true" {
		log.Printf("INFO: %s is not published", workID)
		c.String(http.StatusConflict, fmt.Sprintf("%s is not published", workID))
		return
	}

	svc.auditPublicationChange(claims.ComputeID, tgtObj, false)

	fields["draft"] = "true"
	delete(fields, "publish-date")
	_, err = svc.EasyStore.ObjectUpdate(tgtObj, uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: unpublish %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("unpublish failed: %s", err.Error()))
		return
	}
	svc.publishEvent(uvalibrabus.EventWorkUnpublish, svc.Namespace, tgtObj.Id())
	c.String(http.StatusOK, "unpublished")
}

func (svc *serviceContext) adminDeleteWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get %s work %s for deletion", svc.Namespace, workID)
	delObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.BaseComponent)
	if err != nil {
		log.Printf("ERROR: unablle to get  %s work %s: %s", svc.Namespace, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: delete %s work %s", svc.Namespace, workID)
	_, err = svc.EasyStore.ObjectDelete(delObj, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unablle to delete  %s work %s: %s", svc.Namespace, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) adminUpdateMimeTypes(c *gin.Context) {
	var req []string
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("INFO: invalid request for update mime types: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	added := make([]string, 0)
	for _, mt := range req {
		if slices.Index(svc.MimeTypes, mt) == -1 {
			added = append(added, mt)
		}
	}
	if len(added) > 0 {
		log.Printf("INFO: add new mimetypes %s", added)
		if err := svc.DB.Exec("insert into mime_types (mime_type) values ?", added).Error; err != nil {
			log.Printf("ERROR: unable to add new mimetypes %s: %s", added, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	removed := make([]string, 0)
	for _, mt := range svc.MimeTypes {
		if slices.Index(req, mt) == -1 {
			removed = append(removed, mt)
		}
	}
	if len(removed) > 0 {
		log.Printf("INFO: delete removed mimetypes %s", removed)
		if err := svc.DB.Exec("delete from mime_types where mime_type in ?", removed).Error; err != nil {
			log.Printf("ERROR: unable to remove mimetypes %s: %s", removed, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	svc.MimeTypes = req
	log.Printf("INFO: new list of supported mime types %v", req)
	c.JSON(http.StatusOK, req)
}

func (svc *serviceContext) replaceFile(c *gin.Context) {
	workID := c.Param("id")
	fileName := c.Param("name")
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("INFO: unable to get multipart form for file replace %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	formFile := form.File["file"][0]
	log.Printf("INFO: received request to replace file %s from work %s", fileName, workID)

	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.Files)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s for file replace: %s", svc.Namespace, workID, err.Error())
		if strings.Contains(err.Error(), "not exist") {
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	src, err := formFile.Open()
	if err != nil {
		log.Printf("ERROR: unable to open uploaded file: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		log.Printf("ERROR: unable to read upload file %s: %s", formFile.Filename, err.Error())
		return
	}

	mimeType := http.DetectContentType(fileBytes)
	esBlob := uvaeasystore.NewEasyStoreBlob(fileName, mimeType, fileBytes)
	if err := svc.EasyStore.FileUpdate(svc.Namespace, workID, esBlob); err != nil {
		log.Printf("ERROR: unable to update file  %s: %s", fileName, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// NOTE: this call has already been thru user or admin middleware, so claims will be present
	claims := getJWTClaims(c)
	svc.auditFileReplace(claims.ComputeID, tgtObj, fileName)

	c.String(http.StatusOK, "replaced")
}
