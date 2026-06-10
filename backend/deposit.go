package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

// Current set of field names used by libra/easystore:
// "admin-notes": any related administrator notes.
// "author": work author, often the depositor too.
// "create-date": timestamp, date the work was created.
// "modify-date": timestamp, date the work was modified.
// "depositor": work depositor, not necessarily the author.
// "doi": the work DOI/permanent resource link.
// "disposition": indicates work disposition, currently only "imported" to reflect work imported from the existing Libra repository.
// "draft": values "true" or "false" to indicate if a work is a draft or if it has been published
// "embargo-release": timestamp, when embargo expires (if appropriate).
// "embargo-release-visibility": visibility after embargo expires (if appropriate).
// "invitation-sent": timestamp that indicates when an "invitation to deposit" email was sent.
// "publish-date": timestamp when an OA work goes from private to public or when a user clicks Publish on an ETD work.
// "resource-type": type of work, libraOpen only.
// "registrar": the person creating the optional deposit registration
// "sis-sent": timestamp, indicates notification sent to SIS. LibraETD only.
// "source-id": source of the thesis, a unique SIS or Optional identifier. LibraETD only.
// "source": source of the thesis, "sis" or "optional".
// "submitted-sent" timestamp that indicates when a "successfully submitted" email was sent.
// "visibility": work visibility, either "open", "uva" or "embargo".

type updateSettings struct {
	Visibility               string `json:"visibility"`
	EmbargoReleaseDate       string `json:"embargoReleaseDate,omitempty"`
	EmbargoReleaseVisibility string `json:"embargoReleaseVisibility,omitempty"`
}

type etdUpdateRequest struct {
	Work librametadata.ETDWork `json:"work"`
	updateSettings
}

type registrationRequest struct {
	Program  string `json:"program"`
	Degree   string `json:"degree"`
	Students []struct {
		ComputeID string `json:"computeID"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"students"`
}

func (svc *serviceContext) uploadFile(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: file upload request received for work %s", workID)

	esObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.Files)
	if err != nil {
		log.Printf("ERROR: get work %s for file add failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	mpForm, err := c.MultipartForm()
	if err != nil {
		log.Printf("INFO: unable to get multipart form for file upload: %s", err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("unable to get upload: %s", err.Error()))
		return
	}

	// only 1 file can be uploaded at a time
	formFile := mpForm.File["file"][0]
	log.Printf("INFO: receive submission %s", formFile.Filename)
	uploadSrc, err := formFile.Open()
	if err != nil {
		log.Printf("INFO: unable to open multipart form for file upload: %s", err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("unable to open %s: %s", formFile.Filename, err.Error()))
		return
	}
	defer uploadSrc.Close()
	uploadBytes, err := io.ReadAll(uploadSrc)
	if err != nil {
		log.Printf("INFO: unable to read %s: %s", formFile.Filename, err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("unable to read %s: %s", formFile.Filename, err.Error()))
		return
	}

	mimeType := http.DetectContentType(uploadBytes)
	log.Printf("INFO: create easystore file blob for %s with size %d and mime type %s",
		formFile.Filename, len(uploadBytes), mimeType)
	esBlob := uvaeasystore.NewEasyStoreBlob(formFile.Filename, mimeType, uploadBytes)
	if err := svc.EasyStore.FileCreate(esObj.Namespace(), esObj.Id(), esBlob); err != nil {
		log.Printf("ERROR: unable to add %s to easystore: %s", formFile.Filename, err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("add %s failed: %s", formFile.Filename, err.Error()))
		return
	}

	// NOTE: this call has already been thru user or admin middleware, so claims will be present
	claims := getJWTClaims(c)
	svc.auditFileAdd(claims.ComputeID, esObj, formFile.Filename)

	resp := librametadata.FileData{
		Name:      formFile.Filename,
		MimeType:  mimeType,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) deleteFile(c *gin.Context) {
	workID := c.Param("id")
	delFileName := c.Param("name")
	log.Printf("INFO: request to delete %s from work %s received", workID, delFileName)

	esObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.Files)
	if err != nil {
		log.Printf("ERROR: get work %s for file delete failed: %s", workID, err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if err := svc.EasyStore.FileDelete(esObj.Namespace(), esObj.Id(), delFileName); err != nil {
		log.Printf("ERROR: delete %s from %s failed: %s", delFileName, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// NOTE: this call has already been thru user or admin middleware, so claims will be present
	claims := getJWTClaims(c)
	svc.auditFileDelete(claims.ComputeID, esObj, delFileName)

	c.String(http.StatusOK, "ok")
}
