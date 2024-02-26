package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Files            []string     `json:"files"`
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

	log.Printf("INFO: %+v", req)

	c.String(http.StatusNotImplemented, "boop")
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
