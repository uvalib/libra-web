package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uvalib/easystore/uvaeasystore"
)

type authorData struct {
	ComputeID   string `json:"computeID"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Department  string `json:"department"`
	Institution string `json:"institution"`
}

type oaDepositData struct {
	Token            string       `json:"token"`
	ResourceType     string       `json:"resourceType"`
	Title            string       `json:"title"`
	Authors          []authorData `json:"authors"`
	Abstract         string       `json:"abstract"`
	License          string       `json:"license"`
	Languages        []string     `json:"languages"`
	Keywords         []string     `json:"keywords"`
	Contributors     []authorData `json:"contributors"`
	Published        string       `json:"publisher"`
	Citation         string       `json:"citation"`
	PubllicationData string       `json:"pubDate"`
	RelatedURLs      []string     `json:"relatedURLs"`
	Sponsors         []string     `json:"sponsors"`
	Notes            string       `json:"notes"`
}

type easystoreOAWrapper struct {
	JSONData   oaDepositData
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func (oa easystoreOAWrapper) MimeType() string {
	return "application/json"
}

func (oa easystoreOAWrapper) Payload() []byte {
	out, _ := json.Marshal(oa.JSONData)
	return out
}

func (oa easystoreOAWrapper) PayloadNative() []byte {
	out, _ := json.Marshal(oa.JSONData)
	return out
}

func (oa easystoreOAWrapper) Created() time.Time {
	return oa.CreatedAt
}

func (oa easystoreOAWrapper) Modified() time.Time {
	return oa.ModifiedAt
}

func (svc *serviceContext) getDepositToken(c *gin.Context) {
	log.Printf("INFO: request a deposit token")
	newUUID := uuid.New()
	c.String(http.StatusOK, newUUID.String())
}

func (svc *serviceContext) oaSubmit(c *gin.Context) {
	log.Printf("INFO: received oa deposit request")
	var req oaDepositData
	err := c.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("ERROR: bad payload in oa deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := path.Join("/tmp", req.Token)
	log.Printf("INFO: uploaded subission file location: %s", uploadDir)

	oID := fmt.Sprintf("oid:%s", req.Token)
	obj := uvaeasystore.NewEasyStoreObject(oID)

	md := easystoreOAWrapper{JSONData: req, CreatedAt: time.Now()}
	obj.SetMetadata(md)

	log.Printf("INFO: add files associated with submission token %s", req.Token)
	esFiles := make([]uvaeasystore.EasyStoreBlob, 0)
	err = filepath.Walk(uploadDir, func(fullPath string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ERROR: directory %s traverse failed: %s", uploadDir, err.Error())
			return nil
		}
		if f.IsDir() {
			log.Printf("INFO: directory %s", f.Name())
			return nil
		}

		log.Printf("INFO: add %s", fullPath)
		fileBytes, fileErr := os.ReadFile(fullPath)
		if fileErr != nil {
			return fileErr
		}
		mimeType := http.DetectContentType(fileBytes)
		esBlob := uvaeasystore.NewEasyStoreBlob(f.Name(), mimeType, fileBytes)
		esFiles = append(esFiles, esBlob)
		return nil
	})
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	log.Printf("INFO: cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	obj.SetFiles(esFiles)

	_, err = svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore create failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, oID)
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
	if pathExists(uploadDir) {
		log.Printf("INFO; upload dir %s already exists; cleaning up", uploadDir)
		err = os.RemoveAll(uploadDir)
		if err != nil {
			log.Printf("ERROR: unable to remove %s: %s", uploadDir, err.Error())
		}
	}

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
