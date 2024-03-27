package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

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
	etdWork, err := librametadata.ETDWorkFromBytes(mdBytes)
	if err != nil {
		log.Printf("ERROR: unable to process paypad from work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// enforce work visibility;
	//   values: open (anyone), uva (UVA only), draft: authors only
	visibility := tgtObj.Fields()["default-visibility"]
	depositor := tgtObj.Fields()["depositor"]
	isDraft, err := strconv.ParseBool(tgtObj.Fields()["draft"])
	if err != nil {
		log.Printf("ERROR: unable to parse draft field for etd work %s; default to true: %s", workID, err.Error())
		isDraft = true
	}
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		accessOK := false
		if isDraft {
			accessOK = etdWork.IsAuthor(jwt.ComputeID) || depositor == jwt.ComputeID
		} else {
			if visibility == "open" || etdWork.IsAuthor(jwt.ComputeID) || depositor == jwt.ComputeID {
				accessOK = true
			} else if visibility == "uva" {
				accessOK = svc.isFromUVA(c)
			}
		}
		if accessOK == false {
			log.Printf("INFO: authenticated request by %s to restricted content %s", jwt.ComputeID, workID)
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
	} else {
		if isDraft {
			log.Printf("INFO: non-authenticated request to draft content %s", workID)
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
		if visibility == "uva" {
			if svc.isFromUVA(c) == false {
				log.Printf("INFO: non-authenticated request to restricted etd content %s", workID)
				c.String(http.StatusForbidden, "access to %s is not authorized", workID)
				return
			}
		}
	}

	resp := etdWorkDetails{
		ID:             tgtObj.Id(),
		Version:        tgtObj.VTag(),
		ETDWork:        etdWork,
		IsDraft:        isDraft,
		Visibility:     tgtObj.Fields()["default-visibility"],
		PersistentLink: tgtObj.Fields()["doi"],
		CreatedAt:      tgtObj.Created(),
		ModifiedAt:     tgtObj.Modified()}
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
	oaWork, err := librametadata.OAWorkFromBytes(mdBytes)
	if err != nil {
		log.Printf("ERROR: unable to process paypad from work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// enforce work visibility;
	//   values: open (anyone), authenticated (UVA only), restricted (private; owner only), embargo (owner only)
	visibility := calculateVisibility(tgtObj.Fields())
	depositor := tgtObj.Fields()["depositor"]
	log.Printf("INFO: enforce access to %s with visibility %s", workID, visibility)
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		accessOK := false
		if visibility == "open" || oaWork.IsAuthor(jwt.ComputeID) || depositor == jwt.ComputeID {
			accessOK = true
		} else if visibility == "uva" {
			accessOK = svc.isFromUVA(c)
		}
		if accessOK == false {
			log.Printf("INFO: authenticated request by %s to restricted oa content %s", jwt.ComputeID, workID)
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
	} else {
		log.Printf("INFO: access is from a non-authenticated user")
		if visibility == "uva" {
			if svc.isFromUVA(c) == false {
				log.Printf("INFO: non-authenticated request to restricted oa content %s", workID)
				c.String(http.StatusForbidden, "access to %s is not authorized", workID)
				return
			}
		} else if visibility == "restricted" {
			log.Printf("INFO: non-authenticated request to restricted oa content %s", workID)
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
	}

	resp := oaWorkDetails{ID: tgtObj.Id(),
		Version:        tgtObj.VTag(),
		Visibility:     visibility,
		OAWork:         oaWork,
		PersistentLink: tgtObj.Fields()["doi"],
		CreatedAt:      tgtObj.Created(),
		ModifiedAt:     tgtObj.Modified()}
	if visibility == "embargo" {
		embInfo := embargoData{ReleaseDate: tgtObj.Fields()["embargo-release-date"], ReleaseVisibility: tgtObj.Fields()["embargo-release-visibilty"]}
		resp.Embargo = &embInfo
	}
	for _, oaFile := range tgtObj.Files() {
		log.Printf("INFO: add file %s %s to work", oaFile.Name(), oaFile.Url())
		resp.Files = append(resp.Files, librametadata.FileData{Name: oaFile.Name(), MimeType: oaFile.MimeType(), CreatedAt: oaFile.Created()})
	}

	c.JSON(http.StatusOK, resp)

}
func (svc *serviceContext) isFromUVA(c *gin.Context) bool {
	fromUVA := false
	fwdIP := net.ParseIP(c.Request.Header.Get("X-Forwarded-For"))
	remoteIP := net.ParseIP(c.RemoteIP())
	log.Printf("INFO: check if request is from uva: remote ip: [%s], x-forwarded-for [%s]", c.RemoteIP(), fwdIP)
	for _, ipNet := range svc.UVAWhiteList {
		if ipNet.Contains(fwdIP) || ipNet.Contains(remoteIP) {
			fromUVA = true
			break
		}
	}
	return fromUVA
}

func calculateVisibility(fields uvaeasystore.EasyStoreObjectFields) string {
	visibility := fields["default-visibility"]
	if visibility != "embargo" {
		return visibility
	}
	releaseDateStr := fields["embargo-release"]
	releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		log.Printf("ERROR: unable to parse embardo release date %s: %s", releaseDateStr, err.Error())
		return visibility
	}

	if time.Now().After(releaseDate) {
		return fields["embargo-release-visibility"]
	}
	return visibility
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
	log.Printf("INFO: visibility [%s]", oaSub.Visibility)
	fields := tgtObj.Fields()
	fields["author"] = oaSub.Work.Authors[0].ComputeID
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["default-visibility"] = oaSub.Visibility
	if oaSub.Visibility == "embargo" {
		fields["embargo-release"] = oaSub.EmbargoReleaseDate
		fields["embargo-release-visibility"] = oaSub.EmbargoReleaseVisibility
	}
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)

	resp := oaWorkDetails{ID: tgtObj.Id(),
		Visibility: oaSub.Visibility,
		Version:    updatedObj.VTag(),
		OAWork:     &oaSub.Work,
		CreatedAt:  updatedObj.Created(),
		ModifiedAt: updatedObj.Modified()}
	if oaSub.Visibility == "embargo" {
		embInfo := embargoData{ReleaseDate: oaSub.EmbargoReleaseDate, ReleaseVisibility: oaSub.EmbargoReleaseVisibility}
		resp.Embargo = &embInfo
	}
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
	fields["default-visibility"] = etdReq.Visibility
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)

	resp := etdWorkDetails{ID: tgtObj.Id(),
		Version:    updatedObj.VTag(),
		Visibility: etdReq.Visibility,
		ETDWork:    &etdReq.Work,
		CreatedAt:  updatedObj.Created(),
		ModifiedAt: updatedObj.Modified()}
	for _, file := range updatedObj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
	}
	c.JSON(http.StatusOK, resp)
}
