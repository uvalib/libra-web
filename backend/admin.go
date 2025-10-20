package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

func (svc *serviceContext) adminSearch(c *gin.Context) {
	qStr := c.Query("q")
	if qStr == "" {
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

	log.Printf("INFO: admin search for works with [%s]", qStr)
	payload := map[string]any{"q": qStr, "offset": offset, "limit": limit}
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
	filters := make([]string, 0)
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

	if len(filters) > 0 {
		payload["filter"] = filters
	}

	log.Printf("PAYLOAD %v", payload)
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: search for [%s] failed: %s", qStr, respErr.Message)
		c.String(respErr.StatusCode, respErr.Message)
		return
	}

	// log.Printf("INDEX RESP: %s", rawResp)

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

	// refresh user auth, then get target user details
	err := svc.checkUserServiceJWT()
	if err != nil {
		log.Printf("ERROR: unable to check user service jwt for admin impersonate request: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	url := fmt.Sprintf("%s/user/%s?auth=%s", svc.UserService.URL, tgtComputeID, svc.UserService.JWT)
	resp, userErr := svc.sendGetRequest(url)
	if userErr != nil {
		log.Printf("ERROR: unable get info for user %s impersonate request: %s", tgtComputeID, userErr.Message)
		c.String(userErr.StatusCode, userErr.Message)
		return
	}
	var jsonResp userServiceResp
	err = json.Unmarshal(resp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse user serice response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	jsonResp.User.Role = "user"
	expirationTime := time.Now().Add(1 * time.Hour)
	log.Printf("INFO: generate jwt for impersonated user %+v with expiration %s", jsonResp.User, expirationTime.String())
	claims := jwtClaims{
		UserDetails: &jsonResp.User,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "libra-web",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedStr, jwtErr := token.SignedString([]byte(svc.JWTKey))
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
