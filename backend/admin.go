package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

func (svc *serviceContext) adminExportReport(c *gin.Context) {
	var req struct {
		Q      string `json:"q"`
		Sort   string `json:"sort"`
		Order  string `json:"order"`
		Status string `json:"status"`
		Source string `json:"source"`
		From   string `json:"from"`
		To     string `json:"to"`
		Total  int64  `json:"total"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("INFO: invalid request for export: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: admin export request %+v", req)
	payload := map[string]any{"q": req.Q, "offset": 0, "limit": req.Total}
	if req.Sort != "" {
		switch req.Sort {
		case "created":
			payload["sort"] = []string{fmt.Sprintf("fields.create-date:%s", req.Order)}
		case "title":
			payload["sort"] = []string{fmt.Sprintf("metadata.title:%s", req.Order)}
		case "published":
			payload["sort"] = []string{fmt.Sprintf("fields.publish-date:%s", req.Order)}
		}
	}
	filters := make([]string, 0)
	if req.Source != "any" {
		filters = append(filters, fmt.Sprintf("fields.source=%s", req.Source))
	}
	if req.Status != "any" {
		filters = append(filters, fmt.Sprintf("fields.draft=%t", req.Status == "draft"))
	}
	if req.From != "" {
		dateQ := fmt.Sprintf("fields.publish-date >= %s", req.From)
		if req.To != "" {
			dateQ += fmt.Sprintf(" AND fields.publish-date <= %s", req.To)
		}
		filters = append(filters, dateQ)
	} else if req.To != "" {
		filters = append(filters, fmt.Sprintf("fields.publish-date <= %s", req.To))
	}

	if len(filters) > 0 {
		payload["filter"] = filters
	}

	log.Printf("INFO: export payload %+v", payload)
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: export failed: %s", respErr.Message)
		c.String(respErr.StatusCode, respErr.Message)
		return
	}

	var jsonResp indexResp
	if err := json.Unmarshal(rawResp, &jsonResp); err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

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
		if c.Query("from") != "" {
			dateQ := fmt.Sprintf("fields.publish-date >= %s", c.Query("from"))
			if c.Query("to") != "" {
				dateQ += fmt.Sprintf(" AND fields.publish-date <= %s", c.Query("to"))
			}
			filters = append(filters, dateQ)
		} else if c.Query("to") != "" {
			filters = append(filters, fmt.Sprintf("fields.publish-date <= %s", c.Query("to")))
		}
	}

	if len(filters) > 0 {
		payload["filter"] = filters
	}

	log.Printf("INFO: search payload [%v]", payload)
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

	resp := svc.parseIndexSearchHits(jsonResp)

	log.Printf("INFO: received %d search hits. Elapsed Time: %d (ms)", len(resp.Hits), jsonResp.ProcessingTime)

	c.JSON(http.StatusOK, resp)
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

func (svc *serviceContext) submitOptionalRegistrations(c *gin.Context) {
	var regReq registrationRequest
	err := c.ShouldBindJSON(&regReq)
	if err != nil {
		log.Printf("ERROR: bad payload for optional registration request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// note: this endpoint is protected by admin middleware which ensures claims are present and admin
	claims := getJWTClaims(c)
	log.Printf("INFO: %s requests optional registrations %+v", claims.ComputeID, regReq)
	for _, student := range regReq.Students {
		author := librametadata.ContributorData{ComputeID: student.ComputeID,
			FirstName: student.FirstName, LastName: student.LastName, Institution: "University of Virginia"}
		etdReg := librametadata.ETDWork{Program: regReq.Program, Degree: regReq.Degree, Author: author}
		obj := uvaeasystore.NewEasyStoreObject(svc.Namespace, "")
		fields := uvaeasystore.DefaultEasyStoreFields()
		fields["create-date"] = time.Now().UTC().Format(svc.TimeFormat)
		fields["draft"] = "true"
		fields["default-visibility"] = ""
		fields["depositor"] = student.ComputeID
		fields["registrar"] = claims.ComputeID
		fields["source"] = "optional"
		obj.SetFields(fields)

		// An ETDWork does not serialize the same way as an EasyStoreMetadata object
		// does when being managed by json.Marshal/json.Unmarshal so we wrap it in an object that
		// behaves appropriately
		pl, err := etdReg.Payload()
		if err != nil {
			log.Printf("ERROR: serializing ETDWork: %s", err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		obj.SetMetadata(uvaeasystore.NewEasyStoreMetadata(etdReg.MimeType(), pl))

		_, err = svc.EasyStore.ObjectCreate(obj)
		if err != nil {
			log.Printf("ERROR: admin create registration failed: %s", err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.String(http.StatusOK, fmt.Sprintf("%d registrations completed", len(claims.ComputeID)))
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

type depositStatusResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details []struct {
		ComputingID string `json:"computing_id"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Title       string `json:"title"`
		AcceptedAt  string `json:"accepted_at"`
		ExportedAt  string `json:"exported_at"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"details"`
}

func (svc *serviceContext) adminDepositStatusSearch(c *gin.Context) {
	q := c.Query("q")
	qType := c.Query("type")
	log.Printf("INFO: query deposit status %s=%s", qType, q)
	if err := svc.Protected.refreshJWT(svc.JWTKey); err != nil {
		log.Printf("ERROR: unable to refrest jwt for deposit status search: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := fmt.Sprintf("%s?%s=%s&auth=%s", svc.Protected.DepositAuthURL, qType, q, svc.Protected.JWT)
	respBytes, qErr := svc.sendGetRequest(url)
	if qErr != nil {
		log.Printf("ERROR: deposit status search for %s=%s failed: %s", qType, q, qErr.Message)
		c.String(qErr.StatusCode, qErr.Message)
		return
	}
	var parsed depositStatusResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		log.Printf("ERROR: unable to parse deposit status results: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if !(parsed.Status == 200 || parsed.Status == 201) {
		log.Printf("INFO: got failed %d response to status query %s=%s: %s", parsed.Status, qType, q, parsed.Message)
		c.String(parsed.Status, parsed.Message)
		return
	}

	type hitRec struct {
		ComputeID       string `json:"computeID"`
		FullName        string `json:"fullName"`
		Title           string `json:"title"`
		ReceivedFromSIS string `json:"receivedFromSIS"`
		SumittedToLibra string `json:"submittedToLibra"`
		ExportedToSIS   string `json:"exportedToSIS"`
	}

	/*
		MAPPINGS FROM ORIGINAL RAILS CODE
		CID = deposit['computing_id']
		FULL NAME = "#{deposit['last_name']}, #{deposit['first_name']}"
		RECEIVE FROM SIS = formatted_date( deposit['updated_at'].blank? ? deposit['created_at'] : deposit['updated_at'] )
		SUBMIT TO LIBRA = formatted_date( deposit['accepted_at'] )
		EXPORT TO SIS = deposit['exported_at'].blank? ? 'pending' : formatted_date( deposit['exported_at'] )
		TITLE =  truncate( deposit['title'], length: 75 )
	*/

	var resp = make([]hitRec, 0)
	for _, rawHit := range parsed.Details {
		hit := hitRec{
			ComputeID:       rawHit.ComputingID,
			FullName:        fmt.Sprintf("%s, %s", rawHit.LastName, rawHit.FirstName),
			Title:           rawHit.Title,
			SumittedToLibra: rawHit.AcceptedAt,
		}

		hit.ReceivedFromSIS = rawHit.UpdatedAt
		if hit.ReceivedFromSIS == "" {
			hit.ReceivedFromSIS = rawHit.CreatedAt
		}
		hit.ExportedToSIS = rawHit.ExportedAt
		if hit.ExportedToSIS == "" {
			hit.ExportedToSIS = "pending"
		}
		resp = append(resp, hit)
	}

	c.JSON(http.StatusOK, resp)
}
