package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
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
// "sis-sent": timestamp, indicates notification sent to SIS. LibraETD only.
// "source-id": source of the thesis, a unique SIS or Optional identifier. LibraETD only.
// "source": source of the thesis, "sis" or "optional".
// "submitted-sent" timestamp that indicates when a "successfully submitted" email was sent.
// "visibility": work visibility, either "open", "uva" or "restricted".

type updateSettings struct {
	Visibility               string     `json:"visibility"`
	EmbargoReleaseDate       *time.Time `json:"embargoReleaseDate,omitempty"`
	EmbargoReleaseVisibility string     `json:"embargoReleaseVisibility,omitempty"`
	AddFiles                 []string   `json:"addFiles"`
	DelFiles                 []string   `json:"delFiles"`
}

type oaUpdateRequest struct {
	Work librametadata.OAWork `json:"work"`
	updateSettings
}

type etdUpdateRequest struct {
	Work librametadata.ETDWork `json:"work"`
	updateSettings
}

type registrationRequest struct {
	Department string `json:"department"`
	Degree     string `json:"degree"`
	Students   []struct {
		ComputeID  string `json:"computeID"`
		Email      string `json:"email"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		Department string `json:"department"`
	} `json:"students"`
}

func (svc *serviceContext) getDepositToken(c *gin.Context) {
	log.Printf("INFO: request a deposit token")
	guid := xid.New()
	c.String(http.StatusOK, guid.String())
}

func (svc *serviceContext) adminDepositRegistrations(c *gin.Context) {
	var regReq registrationRequest
	err := c.ShouldBindJSON(&regReq)
	if err != nil {
		log.Printf("ERROR: bad payload for depost registration request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// note: this endpoint is protected by admin middleware which ensures claims are present and admin
	claims := getJWTClaims(c)
	log.Printf("INFO: %s requests deposit registrations %+v", claims.ComputeID, regReq)
	for _, student := range regReq.Students {
		author := librametadata.ContributorData{ComputeID: student.ComputeID,
			FirstName: student.FirstName, LastName: student.LastName, Department: regReq.Department,
			Institution: "University of Virginia"}
		etdReg := librametadata.ETDWork{Program: regReq.Department, Degree: regReq.Degree, Author: author}
		obj := uvaeasystore.NewEasyStoreObject(svc.Namespaces.etd, "")
		fields := uvaeasystore.DefaultEasyStoreFields()
		fields["create-date"] = time.Now().Format(time.RFC3339)
		fields["draft"] = "true"
		fields["default-visibility"] = "open"
		fields["depositor"] = student.ComputeID
		fields["source"] = "optional"
		obj.SetMetadata(etdReg)
		obj.SetFields(fields)

		_, err := svc.EasyStore.Create(obj)
		if err != nil {
			log.Printf("ERROR: admin create registration failed: %s", err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	c.String(http.StatusOK, fmt.Sprintf("%d registrations completed", len(claims.ComputeID)))
}

func (svc *serviceContext) oaDeposit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received oa deposit request for %s", token)
	var oaSub oaUpdateRequest
	err := c.ShouldBindJSON(&oaSub)
	if err != nil {
		log.Printf("ERROR: bad payload in oa deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// must be signed in to do this so there should always be claims
	claims := getJWTClaims(c)
	if claims == nil {
		log.Printf("WARNING: invalid oa dsposit request without claims")
		c.String(http.StatusForbidden, "you do not have permission to make this submission")
		return
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject(svc.Namespaces.oa, "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = claims.ComputeID
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["create-date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = "true"
	fields["default-visibility"] = oaSub.Visibility
	if oaSub.Visibility == "embargo" {
		fields["embargo-release"] = oaSub.EmbargoReleaseDate.Format(time.RFC3339)
		fields["embargo-release-visibility"] = oaSub.EmbargoReleaseVisibility
	}

	// get all submitted files
	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir, oaSub.AddFiles)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	obj.SetMetadata(oaSub.Work)
	obj.SetFiles(esFiles)
	obj.SetFields(fields)

	log.Printf("INFO: save easystore object with namespace %s, id %s", obj.Namespace(), obj.Id())
	newObj, err := svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore save failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create success; cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	resp, err := svc.parseOAWork(newObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse newly deposited oa work: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func getSubmittedFiles(uploadDir string, fileList []string) ([]uvaeasystore.EasyStoreBlob, error) {
	log.Printf("INFO: get files [%v] associated with submission from location %s", fileList, uploadDir)
	esFiles := make([]uvaeasystore.EasyStoreBlob, 0)
	for _, fn := range fileList {
		fullPath := path.Join(uploadDir, fn)
		log.Printf("INFO: add %s", fullPath)
		fileBytes, fileErr := os.ReadFile(fullPath)
		if fileErr != nil {
			return nil, fileErr
		}
		mimeType := http.DetectContentType(fileBytes)
		esBlob := uvaeasystore.NewEasyStoreBlob(fn, mimeType, fileBytes)
		esFiles = append(esFiles, esBlob)
	}
	return esFiles, nil
}

func (svc *serviceContext) cancelSubmission(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: cancel submission %s", token)
	uploadDir := path.Join("/tmp", token)
	if pathExists(uploadDir) {
		err := os.RemoveAll(uploadDir)
		if err != nil {
			log.Printf("ERROR: unable to remove submission upload direcroty %s: %s", uploadDir, err.Error())
		}
	}
	log.Printf("INFO: submission canceled and temporary files cleaned up")
	c.String(http.StatusOK, "cancled")
}

func (svc *serviceContext) removeSubmissionFile(c *gin.Context) {
	tgtFile := fmt.Sprintf("/tmp/%s/%s", c.Param("token"), c.Param("filename"))
	log.Printf("INFO: remove pending submission file %s", tgtFile)
	err := os.Remove(tgtFile)
	if err != nil {
		log.Printf("ERROR: remove %s failed: %s", tgtFile, err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.String(http.StatusOK, "removed")
}

func (svc *serviceContext) uploadSubmissionFiles(c *gin.Context) {
	log.Printf("INFO: file upload request received")
	token := c.Param("token")
	files, err := c.MultipartForm()
	if err != nil {
		log.Printf("ERROR: unable to get upload images: %s", err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("unable to get files: %s", err.Error()))
		return
	}

	uploadDir := path.Join("/tmp", token)
	for _, sf := range files.File["file"] {
		destFile := path.Join(uploadDir, sf.Filename)
		log.Printf("INFO: receive submission to %s", destFile)
		err = c.SaveUploadedFile(sf, destFile)
		if err != nil {
			log.Printf("ERROR: unable to save %s: %s", sf.Filename, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.String(http.StatusOK, "ok")
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
