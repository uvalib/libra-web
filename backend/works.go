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
	"github.com/uvalib/librabus-sdk/uvalibrabus"
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

	// enforce visibility; (admins can always view all)
	//   ANYONE CAN ACCESS METADATA - Except draft content
	//   files: open = visible to all; limited = only visible to uva for a limited timeframe
	visibility := tgtObj.Fields()["default-visibility"]
	depositor := tgtObj.Fields()["depositor"]
	isDraft, err := strconv.ParseBool(tgtObj.Fields()["draft"])
	if err != nil {
		log.Printf("ERROR: unable to parse draft field for etd work %s; default to true: %s", workID, err.Error())
		isDraft = true
	}
	canAccessFiles := false
	canAccessMetadata := true
	isAuthor := false
	isAdmin := false
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		isAuthor = etdWork.IsAuthor(jwt.ComputeID) || depositor == jwt.ComputeID
		isAdmin = jwt.IsAdmin
	}

	if isAuthor || isAdmin {
		canAccessFiles = true
		canAccessMetadata = true
	} else {
		if isDraft {
			canAccessMetadata = false
		} else {
			if visibility == "open" {
				canAccessFiles = true
			} else {
				canAccessFiles = svc.isFromUVA(c)
			}
		}
		if canAccessMetadata == false {
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
	}

	resp := etdWorkDetails{
		baseWorkDetails: &baseWorkDetails{
			ID:             tgtObj.Id(),
			IsDraft:        isDraft,
			Version:        tgtObj.VTag(),
			Visibility:     visibility,
			PersistentLink: tgtObj.Fields()["doi"],
			Files:          make([]librametadata.FileData, 0),
			CreatedAt:      tgtObj.Created(),
			ModifiedAt:     tgtObj.Modified(),
		},
		ETDWork: etdWork,
	}
	if visibility == "uva" {
		resp.Embargo = &embargoData{ReleaseDate: tgtObj.Fields()["embargo-release"], ReleaseVisibility: tgtObj.Fields()["embargo-release-visibility"]}
	}
	if isDraft == false {
		pubDateStr := tgtObj.Fields()["publish-date"]
		pubDate, _ := time.Parse(time.RFC3339, pubDateStr)
		resp.DatePublished = &pubDate
	}
	if canAccessFiles {
		for _, etdFile := range tgtObj.Files() {
			log.Printf("INFO: add file %s %s to work", etdFile.Name(), etdFile.Url())
			resp.Files = append(resp.Files, librametadata.FileData{Name: etdFile.Name(), MimeType: etdFile.MimeType(), CreatedAt: etdFile.Created()})
		}
	} else {
		log.Printf("INFO: access to files for work %s is restricted", workID)
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

	// enforce work visibility; (admin can see all)
	//   ANYONE CAN ACCESS METADATA - Except restricted (private) content
	//   open: anyone can access files
	//   uva: uva users can access files
	//   restricted: only authors can access metadata/files
	//   embargo: files blocked to all but owner
	visibility := calculateVisibility(tgtObj.Fields())
	depositor := tgtObj.Fields()["depositor"]
	log.Printf("INFO: enforce access to %s with visibility %s", workID, visibility)
	canAccessFiles := false
	canAccessMetadata := true
	fromUVA := svc.isFromUVA(c)
	isAuthor := false
	isAdmin := false
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		isAuthor = oaWork.IsAuthor(jwt.ComputeID) || depositor == jwt.ComputeID
		isAdmin = jwt.IsAdmin
	}

	if isAdmin || isAuthor {
		canAccessFiles = true
		canAccessMetadata = true
	} else {
		if visibility == "open" {
			canAccessFiles = true
		} else if visibility == "uva" {
			canAccessFiles = fromUVA
		}
		if canAccessMetadata == false {
			c.String(http.StatusForbidden, "access to %s is not authorized", workID)
			return
		}
	}

	pubDateStr, hasPublicDate := tgtObj.Fields()["publish-date"]
	resp := oaWorkDetails{
		baseWorkDetails: &baseWorkDetails{
			ID:             tgtObj.Id(),
			IsDraft:        !hasPublicDate,
			Version:        tgtObj.VTag(),
			Visibility:     visibility,
			PersistentLink: tgtObj.Fields()["doi"],
			Files:          make([]librametadata.FileData, 0),
			CreatedAt:      tgtObj.Created(),
			ModifiedAt:     tgtObj.Modified(),
		},
		OAWork: oaWork,
	}
	if hasPublicDate {
		pubDate, _ := time.Parse(time.RFC3339, pubDateStr)
		resp.DatePublished = &pubDate
	}
	if visibility == "embargo" {
		embInfo := embargoData{ReleaseDate: tgtObj.Fields()["embargo-release"], ReleaseVisibility: tgtObj.Fields()["embargo-release-visibility"]}
		resp.Embargo = &embInfo
	}

	if canAccessFiles {
		for _, oaFile := range tgtObj.Files() {
			log.Printf("INFO: add file %s %s to work", oaFile.Name(), oaFile.Url())
			resp.Files = append(resp.Files, librametadata.FileData{Name: oaFile.Name(), MimeType: oaFile.MimeType(), CreatedAt: oaFile.Created()})
		}
	} else {
		log.Printf("INFO: access to files for work %s is restricted", workID)
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
	var oaSub oaDepositRequest
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
	publishDateStr, hasPublicDate := fields["publish-date"]
	var publishDate time.Time
	fields["author"] = oaSub.Work.Authors[0].ComputeID
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["default-visibility"] = oaSub.Visibility
	if oaSub.Visibility == "embargo" {
		fields["embargo-release"] = oaSub.EmbargoReleaseDate
		fields["embargo-release-visibility"] = oaSub.EmbargoReleaseVisibility
	}

	visibility := calculateVisibility(tgtObj.Fields())
	fields["draft"] = "true"
	if visibility != "restricted" {
		fields["draft"] = "false"
		if hasPublicDate == false {
			publishDate = time.Now()
			fields["publish-date"] = publishDate.Format(time.RFC3339)
			hasPublicDate = true
			svc.publishEvent(uvalibrabus.EventWorkPublish, svc.Namespaces.oa, tgtObj.Id())
		} else {
			publishDate, _ = time.Parse(time.RFC3339, publishDateStr)
		}
	}

	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)
	resp := oaWorkDetails{
		baseWorkDetails: &baseWorkDetails{
			ID:             updatedObj.Id(),
			IsDraft:        !hasPublicDate,
			Version:        updatedObj.VTag(),
			Visibility:     oaSub.Visibility,
			PersistentLink: updatedObj.Fields()["doi"],
			CreatedAt:      updatedObj.Created(),
			ModifiedAt:     updatedObj.Modified(),
		},
		OAWork: &oaSub.Work,
	}
	if oaSub.Visibility == "embargo" {
		embInfo := embargoData{ReleaseDate: oaSub.EmbargoReleaseDate, ReleaseVisibility: oaSub.EmbargoReleaseVisibility}
		resp.Embargo = &embInfo
	}
	if publishDate.IsZero() == false {
		resp.DatePublished = &publishDate
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
	var etdReq etdDepositRequest
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
	if etdReq.Visibility == "uva" {
		fields["embargo-release"] = etdReq.EmbargoReleaseDate
		fields["embargo-release-visibility"] = etdReq.EmbargoReleaseVisibility
	}
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)
	isDraft, _ := strconv.ParseBool(updatedObj.Fields()["draft"])
	resp := etdWorkDetails{
		baseWorkDetails: &baseWorkDetails{
			ID:             updatedObj.Id(),
			IsDraft:        isDraft,
			Version:        updatedObj.VTag(),
			Visibility:     etdReq.Visibility,
			PersistentLink: updatedObj.Fields()["doi"],
			CreatedAt:      updatedObj.Created(),
			ModifiedAt:     updatedObj.Modified(),
		},
		ETDWork: &etdReq.Work,
	}
	if etdReq.Visibility == "uva" {
		resp.Embargo = &embargoData{ReleaseDate: etdReq.EmbargoReleaseDate, ReleaseVisibility: etdReq.EmbargoReleaseVisibility}
	}
	if isDraft == false {
		pubDateStr := tgtObj.Fields()["publish-date"]
		pubDate, _ := time.Parse(time.RFC3339, pubDateStr)
		resp.DatePublished = &pubDate
	}

	for _, file := range updatedObj.Files() {
		resp.Files = append(resp.Files, librametadata.FileData{MimeType: file.MimeType(), Name: file.Name(), CreatedAt: file.Created()})
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) publishETDWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get etd work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.etd, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespaces.etd, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	fields := tgtObj.Fields()
	fields["draft"] = "false"
	fields["publish-date"] = time.Now().Format(time.RFC3339)
	_, err = svc.EasyStore.Update(tgtObj, uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: unable ti publish etd work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	svc.publishEvent(uvalibrabus.EventWorkPublish, svc.Namespaces.etd, tgtObj.Id())
	c.String(http.StatusOK, "published")
}
