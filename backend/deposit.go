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

type oaWorkRequest struct {
	Work     librametadata.OAWork `json:"work"`
	AddFiles []string             `json:"addFiles"`
	DelFiles []string             `json:"delFiles"`
}

type etdWorkRequest struct {
	Work     librametadata.ETDWork `json:"work"`
	AddFiles []string              `json:"addFiles"`
	DelFiles []string              `json:"delFiles"`
}

type versionedOA struct {
	ID         string                   `json:"id"`
	Version    string                   `json:"version"`
	CreatedAt  time.Time                `json:"createdAt"`
	ModifiedAt time.Time                `json:"modifiedAt"`
	Files      []librametadata.FileData `json:"files"`
	*librametadata.OAWork
}

type versionedETD struct {
	ID         string                   `json:"id"`
	Version    string                   `json:"version"`
	CreatedAt  time.Time                `json:"createdAt"`
	ModifiedAt time.Time                `json:"modifiedAt"`
	Files      []librametadata.FileData `json:"files"`
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
	var etdSub etdWorkRequest
	err := c.ShouldBindJSON(&etdSub)
	if err != nil {
		log.Printf("ERROR: bad payload in etd deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir, etdSub.AddFiles)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject(svc.Namespaces.etd, "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = etdSub.Work.Author.ComputeID
	fields["title"] = etdSub.Work.Title
	obj.SetMetadata(etdSub.Work)
	obj.SetFiles(esFiles)
	obj.SetFields(fields)

	log.Printf("INFO: save easystore object with namespace %s, id %s", obj.Namespace(), obj.Id())
	_, err = svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore save failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create success; cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	resp := versionedETD{ID: obj.Id(), Version: obj.VTag(), ETDWork: &etdSub.Work, CreatedAt: obj.Created(), ModifiedAt: obj.Modified()}
	for _, file := range obj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) oaSubmit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received oa deposit request for %s", token)
	var oaSub oaWorkRequest
	err := c.ShouldBindJSON(&oaSub)
	if err != nil {
		log.Printf("ERROR: bad payload in oa deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir, oaSub.AddFiles)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject(svc.Namespaces.oa, "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = oaSub.Work.Authors[0].ComputeID
	fields["title"] = oaSub.Work.Title
	fields["publisher"] = oaSub.Work.Publisher
	fields["resourceType"] = oaSub.Work.ResourceType
	obj.SetMetadata(oaSub.Work)
	obj.SetFiles(esFiles)
	obj.SetFields(fields)

	log.Printf("INFO: save easystore object with namespace %s, id %s", obj.Namespace(), obj.Id())
	_, err = svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore save failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create success; cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	resp := versionedOA{ID: obj.Id(), Version: obj.VTag(), OAWork: &oaSub.Work, CreatedAt: obj.Created(), ModifiedAt: obj.Modified()}
	for _, file := range obj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
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
