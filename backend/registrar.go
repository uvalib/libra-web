package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

type registrationStudent struct {
	ID             uint64     `json:"-"`
	RegistrationID uint64     `json:"-"`
	WorkID         string     `json:"workID"`
	ComputeID      string     `json:"computeID"`
	CompletedAt    *time.Time `json:"completedAt"`
}

type registration struct {
	ID          uint64                `json:"id"`
	Registrar   string                `json:"registrar"`
	Degree      string                `json:"degree"`
	Program     string                `json:"program"`
	Students    []registrationStudent `json:"students"`
	SubmittedAt time.Time             `json:"submittedAt"`
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

func (svc *serviceContext) sisDepositStatusSearch(c *gin.Context) {
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

type optSearchResults struct {
	Hits  []registration `json:"hits"`
	Total int64          `json:"total"`
}

func (svc *serviceContext) optionalDepositStatusSearch(c *gin.Context) {
	offset, pageErr := strconv.ParseInt(c.Query("offset"), 10, 0)
	if pageErr != nil {
		offset = 0
	}
	limit, pageErr := strconv.ParseInt(c.Query("limit"), 10, 0)
	if pageErr != nil {
		limit = 25
	}
	sort := c.Query("sort")
	if len(sort) == 0 {
		sort = "submitted_at"
	}
	order := c.Query("order")
	if len(sort) == 0 {
		sort = "desc"
	}

	baseQ := svc.DB.Table("registrations")
	if c.Query("registrar") != "" {
		baseQ = baseQ.Where("registrar ~* ?", fmt.Sprintf("^\\s*%s", c.Query("registrar")))
	}
	if c.Query("program") != "" {
		baseQ = baseQ.Where("program = ?", c.Query("program"))
	}
	if c.Query("degree") != "" {
		baseQ = baseQ.Where("degree = ?", c.Query("degree"))
	}
	if c.Query("submitted_at") != "" {
		baseQ = baseQ.Where("submitted_at >= ?", c.Query("submitted_at"))
	}

	log.Printf("INFO: search %d limit %d optional deposits order %s %s", offset, limit, sort, order)
	resp := optSearchResults{}
	if err := baseQ.Count(&resp.Total).Error; err != nil {
		log.Printf("ERROR: unable to get optional registrations count: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := baseQ.Preload("Students").Offset(int(offset)).Limit(int(limit)).Order(fmt.Sprintf("%s %s", sort, order)).Find(&resp.Hits).Error; err != nil {
		log.Printf("ERROR: unable to get optional registrations: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
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

	log.Printf("INFO: create registration record to track status")
	newRegistration := registration{
		Registrar:   claims.ComputeID,
		Degree:      regReq.Degree,
		Program:     regReq.Program,
		SubmittedAt: time.Now(),
	}
	if err := svc.DB.Create(&newRegistration).Error; err != nil {
		log.Printf("ERROR: unable to create registration record: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

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

		log.Printf("INFO: add new work to registrations record %d", newRegistration.ID)
		rec := registrationStudent{
			RegistrationID: newRegistration.ID,
			WorkID:         obj.Id(),
			ComputeID:      student.ComputeID,
		}
		if err := svc.DB.Create(&rec).Error; err != nil {
			log.Printf("ERROR: unable to create student %s work %s registration record for registration %d: %s",
				student.ComputeID, obj.Id(), newRegistration.ID, err.Error())
		}
	}
	c.String(http.StatusOK, fmt.Sprintf("%d registrations completed", len(claims.ComputeID)))
}
