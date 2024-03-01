package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

type savedOA struct {
	ID string `json:"id"`
	librametadata.OAWork
}
type savedETD struct {
	ID string `json:"id"`
	librametadata.ETDWork
}

func (svc *serviceContext) getDepositToken(c *gin.Context) {
	log.Printf("INFO: request a deposit token")
	guid := xid.New()
	c.String(http.StatusOK, guid.String())
}

func (svc *serviceContext) deleteWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}

func (svc *serviceContext) oaUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}

func (svc *serviceContext) etdSubmit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received etd deposit request for %s", token)
	var etdWork librametadata.ETDWork
	err := c.ShouldBindJSON(&etdWork)
	if err != nil {
		log.Printf("ERROR: bad payload in etd deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject("etd", "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = etdWork.Author.ComputeID
	fields["title"] = etdWork.Title
	obj.SetMetadata(etdWork)
	obj.SetFiles(esFiles)
	obj.SetFields(fields)

	log.Printf("INFO: save easystore object with namespace %s, id %s", obj.Id(), obj.Namespace())
	_, err = svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore save failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create success; cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	c.JSON(http.StatusOK, savedETD{ID: obj.AccessId(), ETDWork: etdWork})
}

func (svc *serviceContext) oaSubmit(c *gin.Context) {
	token := c.Param("token")
	log.Printf("INFO: received oa deposit request for %s", token)
	var oaW librametadata.OAWork
	err := c.ShouldBindJSON(&oaW)
	if err != nil {
		log.Printf("ERROR: bad payload in oa deposit request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := path.Join("/tmp", token)
	esFiles, err := getSubmittedFiles(uploadDir)
	if err != nil {
		log.Printf("ERROR: unable to add files from %s: %s", uploadDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("INFO: create easystore object")
	obj := uvaeasystore.NewEasyStoreObject("oa", "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = oaW.Authors[0].ComputeID
	fields["title"] = oaW.Title
	fields["publisher"] = oaW.Publisher
	fields["resourceType"] = oaW.ResourceType
	obj.SetMetadata(oaW)
	obj.SetFiles(esFiles)
	obj.SetFields(fields)

	log.Printf("INFO: save easystore object with namespace %s, id %s", obj.Id(), obj.Namespace())
	_, err = svc.EasyStore.Create(obj)
	if err != nil {
		log.Printf("ERROR: easystore save failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create success; cleanup upload directory %s", uploadDir)
	os.RemoveAll(uploadDir)

	c.JSON(http.StatusOK, savedOA{ID: obj.AccessId(), OAWork: oaW})
}

func getSubmittedFiles(uploadDir string) ([]uvaeasystore.EasyStoreBlob, error) {
	log.Printf("INFO: get files associated with submission from location %s", uploadDir)
	esFiles := make([]uvaeasystore.EasyStoreBlob, 0)
	err := filepath.Walk(uploadDir, func(fullPath string, f os.FileInfo, err error) error {
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
		return make([]uvaeasystore.EasyStoreBlob, 0), err
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
