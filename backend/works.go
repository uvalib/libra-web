package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

func (svc *serviceContext) getETDWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get etd work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.etd, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespaces.etd, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		log.Printf("ERROR: unable to get metadata paload from respose: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	parsedETDWork, err := librametadata.ETDWorkFromBytes(mdBytes)
	if err != nil {
		log.Printf("ERROR: unable to process paypad from work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := versionedETD{ID: tgtObj.Id(), Version: tgtObj.VTag(), ETDWork: parsedETDWork, CreatedAt: tgtObj.Created(), ModifiedAt: tgtObj.Modified()}
	for _, etdFile := range tgtObj.Files() {
		log.Printf("INFO: add file %s %s to work", etdFile.Name(), etdFile.Url())
		resp.Files = append(resp.Files, librametadata.FileData{Name: etdFile.Name(), MimeType: etdFile.MimeType(), CreatedAt: etdFile.Created()})
	}

	c.JSON(http.StatusOK, resp)

}

func (svc *serviceContext) getOAWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get oa work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.oa, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespaces.oa, workID, err.Error())
		if strings.Contains(err.Error(), "not exist") {
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		log.Printf("ERROR: unable to get metadata paload from respose: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	parsedOAWork, err := librametadata.OAWorkFromBytes(mdBytes)
	if err != nil {
		log.Printf("ERROR: unable to process paypad from work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := versionedOA{ID: tgtObj.Id(), Version: tgtObj.VTag(), OAWork: parsedOAWork, CreatedAt: tgtObj.Created(), ModifiedAt: tgtObj.Modified()}
	for _, oaFile := range tgtObj.Files() {
		log.Printf("INFO: add file %s %s to work", oaFile.Name(), oaFile.Url())
		resp.Files = append(resp.Files, librametadata.FileData{Name: oaFile.Name(), MimeType: oaFile.MimeType(), CreatedAt: oaFile.Created()})
	}

	c.JSON(http.StatusOK, resp)

}

func (svc *serviceContext) deleteOAWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete oa work %s", workID)
	err := svc.deleteWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: unable to delete oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) deleteETDWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete oa work %s", workID)
	err := svc.deleteWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: unable to delete oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) deleteWork(namespace string, id string) error {
	log.Printf("INFO: get %s work %s for deletion", namespace, id)
	delObj, err := svc.EasyStore.GetByKey(namespace, id, uvaeasystore.BaseComponent)
	if err != nil {
		return fmt.Errorf("unable to get %s work %s: %s", namespace, id, err.Error())
	}

	log.Printf("INFO: delete %s work %s", namespace, id)
	_, err = svc.EasyStore.Delete(delObj, uvaeasystore.AllComponents)
	return err
}

func (svc *serviceContext) oaUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update oa work %s", workID)
	var oaSub oaWorkRequest
	err := c.ShouldBindJSON(&oaSub)
	if err != nil {
		log.Printf("ERROR: bad payload in oa update request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: load existing oa work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.oa, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: get oa work %s for update failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// update the metadata with newly submitted info
	tgtObj.SetMetadata(oaSub.Work)

	// get a list of the files currently  attached to the work and remove those that have been deleted
	esFiles := tgtObj.Files()
	for _, fn := range oaSub.DelFiles {
		for fIdx, origF := range esFiles {
			if origF.Name() == fn {
				log.Printf("INFO: remove file %s from oa work %s", fn, workID)
				esFiles = append(esFiles[:fIdx], esFiles[fIdx+1:]...)
				break
			}
		}
	}

	// newly uploaded files are in a tmp dir named by the ID of the OA work
	if len(oaSub.AddFiles) > 0 {
		log.Printf("INFO: adding %v to oa work %s", oaSub.AddFiles, workID)
		uploadDir := path.Join("/tmp", workID)
		addedFiles, err := getSubmittedFiles(uploadDir, oaSub.AddFiles)
		if err != nil {
			log.Printf("ERROR: unable to get newly uploaded files from %s: %s", uploadDir, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		esFiles = append(esFiles, addedFiles...)
		log.Printf("INFO: updated files %+v", esFiles)
		tgtObj.SetFiles(esFiles)
		log.Printf("INFO: cleanup upload directory %s", uploadDir)
		os.RemoveAll(uploadDir)
	}

	// update fields
	fields := tgtObj.Fields()
	fields["author"] = oaSub.Work.Authors[0].ComputeID
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["visibility"] = oaSub.Work.Visibility
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)

	resp := versionedOA{ID: tgtObj.Id(), Version: updatedObj.VTag(), OAWork: &oaSub.Work, CreatedAt: updatedObj.Created(), ModifiedAt: updatedObj.Modified()}
	for _, file := range updatedObj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) downloadETDFile(c *gin.Context) {
	workID := c.Param("id")
	tgtFile := c.Param("name")
	log.Printf("INFO: request to download file %s from etd work %s", tgtFile, workID)
	svc.doFileDownload(c, svc.Namespaces.etd, workID, tgtFile)
}

func (svc *serviceContext) downloadOAFile(c *gin.Context) {
	workID := c.Param("id")
	tgtFile := c.Param("name")
	log.Printf("INFO: request to download file %s from oa work %s", tgtFile, workID)
	svc.doFileDownload(c, svc.Namespaces.oa, workID, tgtFile)
}

func (svc *serviceContext) doFileDownload(c *gin.Context, namespace, workID, tgtFile string) {
	log.Printf("INFO: load  %s work %s file info", namespace, workID)
	tgtObj, err := svc.EasyStore.GetByKey(namespace, workID, uvaeasystore.Files)
	if err != nil {
		log.Printf("ERROR: get %s work %s for download failed: %s", namespace, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var dlFile uvaeasystore.EasyStoreBlob
	for _, oaFile := range tgtObj.Files() {
		if oaFile.Name() == tgtFile {
			dlFile = oaFile
		}
	}

	if dlFile == nil {
		log.Printf("INFO: file %s not found in %s work %s", tgtFile, namespace, workID)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", tgtFile))
		return
	}

	bodyBytes, err := dlFile.Payload()
	if err != nil {
		log.Printf("ERROR: unable to get opayload for file %s: %s", tgtFile, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: set %s with mime type %s to client", tgtFile, dlFile.MimeType())
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+tgtFile)
	c.Header("Content-Type", dlFile.MimeType())
	c.Data(http.StatusOK, dlFile.MimeType(), bodyBytes)
}

func (svc *serviceContext) etdUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update etd work %s", workID)
	var etdReq etdWorkRequest
	err := c.ShouldBindJSON(&etdReq)
	if err != nil {
		log.Printf("ERROR: bad payload in etd update request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: load existing oa work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.etd, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: get etd work %s for update failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// update the metadata with newly submitted info
	tgtObj.SetMetadata(etdReq.Work)

	// get a list of the files currently  attached to the work and remove those that have been deleted
	esFiles := tgtObj.Files()
	for _, fn := range etdReq.DelFiles {
		for fIdx, origF := range esFiles {
			if origF.Name() == fn {
				log.Printf("INFO: remove file %s from etd work %s", fn, workID)
				esFiles = append(esFiles[:fIdx], esFiles[fIdx+1:]...)
				break
			}
		}
	}

	// newly uploaded files are in a tmp dir named by the ID of the OA work
	if len(etdReq.AddFiles) > 0 {
		log.Printf("INFO: adding %v to etd work %s", etdReq.AddFiles, workID)
		uploadDir := path.Join("/tmp", workID)
		addedFiles, err := getSubmittedFiles(uploadDir, etdReq.AddFiles)
		if err != nil {
			log.Printf("ERROR: unable to get newly uploaded files from %s: %s", uploadDir, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		esFiles = append(esFiles, addedFiles...)
		log.Printf("INFO: updated files %+v", esFiles)
		tgtObj.SetFiles(esFiles)
		log.Printf("INFO: cleanup upload directory %s", uploadDir)
		os.RemoveAll(uploadDir)
	}

	// update fields
	fields := tgtObj.Fields()
	fields["author"] = etdReq.Work.Author.ComputeID
	fields["visibility"] = etdReq.Work.Visibility
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)

	resp := versionedETD{ID: tgtObj.Id(), Version: updatedObj.VTag(), ETDWork: &etdReq.Work, CreatedAt: updatedObj.Created(), ModifiedAt: updatedObj.Modified()}
	for _, file := range updatedObj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
	}
	c.JSON(http.StatusOK, resp)
}
