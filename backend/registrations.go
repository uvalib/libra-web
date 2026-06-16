package main

import (
	"fmt"
	"log"
	"net/http"
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
	ID          uint64 `json:"-"`
	Registrar   string `json:"registrar"`
	Degree      string `json:"degree"`
	Program     string `json:"program"`
	Students    []registrationStudent
	SubmittedAt time.Time `json:"submittedAt"`
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
