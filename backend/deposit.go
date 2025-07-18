package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
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
// "visibility": work visibility, either "open", "uva" or "restricted".

type updateSettings struct {
	Visibility               string   `json:"visibility"`
	EmbargoReleaseDate       string   `json:"embargoReleaseDate,omitempty"`
	EmbargoReleaseVisibility string   `json:"embargoReleaseVisibility,omitempty"`
	AddFiles                 []string `json:"addFiles"`
	DelFiles                 []string `json:"delFiles"`
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

func (svc *serviceContext) cancelSubmission(c *gin.Context) {
	workID := c.Param("work")
	log.Printf("INFO: cancel submited files for work %s", workID)
	uploadDir := path.Join("/tmp", workID)
	if pathExists(uploadDir) {
		err := os.RemoveAll(uploadDir)
		if err != nil {
			log.Printf("ERROR: unable to remove submission upload direcroty %s: %s", uploadDir, err.Error())
		}
	}
	log.Printf("INFO: temporary submission file %s have been cleaned up", uploadDir)
	c.String(http.StatusOK, "cancled")
}

func (svc *serviceContext) removeSubmissionFile(c *gin.Context) {
	tgtFile := fmt.Sprintf("/tmp/%s/%s", c.Param("work"), c.Param("filename"))
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
	workID := c.Param("work")
	files, err := c.MultipartForm()
	if err != nil {
		log.Printf("ERROR: unable to get upload images: %s", err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("unable to get files: %s", err.Error()))
		return
	}

	uploadDir := path.Join("/tmp", workID)
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
