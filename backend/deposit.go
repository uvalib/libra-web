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

type depositSettings struct {
	Visibility               string   `json:"visibility"`
	EmbargoReleaseDate       string   `json:"embargoReleaseDate,omitempty"`
	EmbargoReleaseVisibility string   `json:"embargoReleaseVisibility,omitempty"`
	AddFiles                 []string `json:"addFiles"`
	DelFiles                 []string `json:"delFiles"`
}

type oaDepositRequest struct {
	Work librametadata.OAWork `json:"work"`
	depositSettings
}

type etdDepositRequest struct {
	Work librametadata.ETDWork `json:"work"`
	depositSettings
}

type embargoData struct {
	ReleaseDate       string `json:"releaseDate"`
	ReleaseVisibility string `json:"releaseVisibility"`
}

type baseWorkDetails struct {
	ID             string                   `json:"id"`
	Version        string                   `json:"version"`
	PersistentLink string                   `json:"persistentLink,omitempty"`
	IsDraft        bool                     `json:"isDraft"`
	Visibility     string                   `json:"visibility"`
	Embargo        *embargoData             `json:"embargo,omitempty"`
	Files          []librametadata.FileData `json:"files"`
	DatePublished  *time.Time               `json:"datePublished,omitempty"`
	CreatedAt      time.Time                `json:"createdAt"`
	ModifiedAt     time.Time                `json:"modifiedAt"`
}

type oaWorkDetails struct {
	*baseWorkDetails
	*librametadata.OAWork
}

type etdWorkDetails struct {
	*baseWorkDetails
	*librametadata.ETDWork
}

func (svc *serviceContext) getDepositToken(c *gin.Context) {
	log.Printf("INFO: request a deposit token")
	guid := xid.New()
	c.String(http.StatusOK, guid.String())
}

func (svc *serviceContext) etdSubmit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received etd deposit request for %s", token)
	var etdSub etdDepositRequest
	err := c.ShouldBindJSON(&etdSub)
	if err != nil {
		log.Printf("ERROR: bad payload in etd deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// must be signed in to do this so there should always be claims
	claims := getJWTClaims(c)
	if claims == nil {
		log.Printf("WARNING: invalid etd dsposit request without claims")
		c.String(http.StatusForbidden, "you do not have permission to make this submission")
		return
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject(svc.Namespaces.etd, "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = claims.Email
	fields["author"] = etdSub.Work.Author.ComputeID
	fields["create-date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = "true"
	fields["default-visibility"] = etdSub.Visibility
	if etdSub.Visibility == "uva" {
		fields["embargo-release"] = etdSub.EmbargoReleaseDate
		fields["embargo-release-visibility"] = etdSub.EmbargoReleaseVisibility
	}

	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir, etdSub.AddFiles)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	obj.SetMetadata(etdSub.Work)
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

	resp, err := parseETDWork(newObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse newly deposited etd work: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) oaSubmit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received oa deposit request for %s", token)
	var oaSub oaDepositRequest
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
	fields["depositor"] = claims.Email
	fields["author"] = oaSub.Work.Authors[0].ComputeID
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["create-date"] = time.Now().Format(time.RFC3339)
	fields["draft"] = "true"
	fields["default-visibility"] = oaSub.Visibility
	if oaSub.Visibility == "embargo" {
		fields["embargo-release"] = oaSub.EmbargoReleaseDate
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

	resp, err := parseOAWork(newObj, true)
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
