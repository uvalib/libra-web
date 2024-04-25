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

type embargoData struct {
	ReleaseDate       *time.Time `json:"releaseDate"`
	ReleaseVisibility string     `json:"releaseVisibility"`
}

type commonWorkDetails struct {
	ID             string                   `json:"id"`
	Version        string                   `json:"version"`
	PersistentLink string                   `json:"persistentLink,omitempty"`
	IsDraft        bool                     `json:"isDraft"`
	Visibility     string                   `json:"visibility"`
	Embargo        *embargoData             `json:"embargo,omitempty"`
	Files          []librametadata.FileData `json:"files"`
	CreatedAt      time.Time                `json:"createdAt"`
	ModifiedAt     *time.Time               `json:"modifiedAt,omitempty"`
	PublishedAt    *time.Time               `json:"publishedAt,omitempty"`
}

func (detail *commonWorkDetails) parseDates(esObj uvaeasystore.EasyStoreObject) {
	detail.CreatedAt = esObj.Created()
	if detail.IsDraft == false {
		pubDateStr, published := esObj.Fields()["publish-date"]
		if published {
			pubDate, err := time.Parse(time.RFC3339, pubDateStr)
			if err != nil {
				log.Printf("ERROR: unable to parse publish-date [%s]: %s", pubDateStr, err.Error())
				pubDate = time.Now()
			}
			detail.PublishedAt = &pubDate
		}
	}
	modDateStr, modified := esObj.Fields()["modify-date"]
	if modified {
		modDate, err := time.Parse(time.RFC3339, modDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse modify-date [%s]: %s", modDateStr, err.Error())
			modDate = time.Now()
		}
		detail.ModifiedAt = &modDate
	}

	embargoDateStr, exist := esObj.Fields()["embargo-release"]
	if exist {
		var releaseDate *time.Time
		parsed, err := time.Parse(time.RFC3339, embargoDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse work %s release date [%s]: %s", esObj.Id(), embargoDateStr, err.Error())
		} else {
			releaseDate = &parsed
		}
		embInfo := embargoData{ReleaseDate: releaseDate, ReleaseVisibility: esObj.Fields()["embargo-release-visibility"]}
		detail.Embargo = &embInfo
	}
}

type oaWorkDetails struct {
	*commonWorkDetails
	*librametadata.OAWork
}

type etdWorkDetails struct {
	*commonWorkDetails
	*librametadata.ETDWork
}

type workAccess struct {
	files    bool
	metadata bool
}

func (svc *serviceContext) getETDWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get etd work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.etd, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespaces.oa, workID, err.Error())
		if strings.Contains(err.Error(), "not exist") {
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.Printf("INFO: check access to ea work %s", workID)
	access := svc.canAccessWork(c, tgtObj)
	if access.metadata == false {
		log.Printf("INFO: access to etd work %s is forbidden", workID)
		c.String(http.StatusForbidden, "access to %s is not authorized", workID)
		return
	}

	etdWork, err := svc.parseETDWork(tgtObj, access.files)
	if err != nil {
		log.Printf("ERROR: unable to parse etd work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, etdWork)
}

func (svc *serviceContext) etdUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update etd work %s", workID)
	var etdReq etdUpdateRequest
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

	canUpdate := false
	claims := getJWTClaims(c)
	if claims != nil {
		depositor := tgtObj.Fields()["depositor"]
		canUpdate = depositor == claims.ComputeID || claims.IsAdmin
	}
	if canUpdate == false {
		log.Printf("INFO: unauthorized attempt to update etd work %s", workID)
		c.String(http.StatusForbidden, "you do not have permission to update this work")
		return
	}

	// update the metadata with newly submitted info
	svc.auditETDWorkUpdate(claims.ComputeID, etdReq, tgtObj)
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
		tgtObj.SetFiles(esFiles)
		log.Printf("INFO: cleanup upload directory %s", uploadDir)
		os.RemoveAll(uploadDir)
	}

	// update fields
	fields := tgtObj.Fields()
	fields["modify-date"] = time.Now().Format(time.RFC3339)
	fields["default-visibility"] = etdReq.Visibility
	if etdReq.Visibility == "uva" {
		if etdReq.EmbargoReleaseDate == nil {
			log.Printf("INFO: etd work %s set for a forever embargo", tgtObj.Id())
			fields["embargo-release"] = ""
			delete(fields, "embargo-release-visibility")
		} else {
			fields["embargo-release"] = etdReq.EmbargoReleaseDate.Format(time.RFC3339)
			fields["embargo-release-visibility"] = etdReq.EmbargoReleaseVisibility
		}
	} else {
		delete(fields, "embargo-release")
		delete(fields, "embargo-release-visibility")
	}
	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)
	resp, err := svc.parseETDWork(updatedObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse updated etd work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) publishETDWork(c *gin.Context) {
	workID := c.Param("id")
	err := svc.publishWork(svc.Namespaces.etd, workID)
	if err != nil {
		log.Printf("ERROR: publish etd work %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "published")
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

	log.Printf("INFO: check access to ea work %s", workID)
	access := svc.canAccessWork(c, tgtObj)
	if access.metadata == false {
		log.Printf("INFO: access to oa work %s is forbidden", workID)
		c.String(http.StatusForbidden, "access to %s is not authorized", workID)
		return
	}

	oaWork, err := svc.parseOAWork(tgtObj, access.files)
	if err != nil {
		log.Printf("ERROR: unable to parse oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, oaWork)
}

func (svc *serviceContext) oaUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update oa work %s", workID)
	var oaSub oaUpdateRequest
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

	canUpdate := false
	claims := getJWTClaims(c)
	if claims != nil {
		depositor := tgtObj.Fields()["depositor"]
		canUpdate = depositor == claims.ComputeID || claims.IsAdmin
	}
	if canUpdate == false {
		log.Printf("INFO: unauthorized attempt to update oa work %s", workID)
		c.String(http.StatusForbidden, "you do not have permission to update this work")
		return
	}

	// update the metadata with newly submitted info
	svc.auditOAWorkUpdate(claims.ComputeID, oaSub, tgtObj)
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
		tgtObj.SetFiles(esFiles)
		log.Printf("INFO: cleanup upload directory %s", uploadDir)
		os.RemoveAll(uploadDir)
	}

	// update fields
	fields := tgtObj.Fields()
	fields["resource-type"] = oaSub.Work.ResourceType
	fields["modify-date"] = time.Now().Format(time.RFC3339)
	fields["default-visibility"] = oaSub.Visibility
	if oaSub.Visibility == "embargo" {
		if oaSub.EmbargoReleaseDate == nil {
			log.Printf("INFO: oa work %s set for a forever embargo", tgtObj.Id())
			fields["embargo-release"] = ""
			delete(fields, "embargo-release-visibility")
		} else {
			fields["embargo-release"] = oaSub.EmbargoReleaseDate.Format(time.RFC3339)
			fields["embargo-release-visibility"] = oaSub.EmbargoReleaseVisibility
		}
	} else {
		delete(fields, "embargo-release")
		delete(fields, "embargo-release-visibility")
	}

	tgtObj.SetFields(fields)

	updatedObj, err := svc.EasyStore.Update(tgtObj, uvaeasystore.AllComponents)
	resp, err := svc.parseOAWork(updatedObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse updated oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) publishOAWork(c *gin.Context) {
	workID := c.Param("id")
	err := svc.publishWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: publish oa work %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "published")
}

func (svc *serviceContext) unpublishOAWork(c *gin.Context) {
	workID := c.Param("id")
	err := svc.unpublishWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: unpublish oa work %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "unpublished")
}

func (svc *serviceContext) unpublishETDWork(c *gin.Context) {
	workID := c.Param("id")
	err := svc.unpublishWork(svc.Namespaces.etd, workID)
	if err != nil {
		log.Printf("ERROR: unpublish etd work %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "unpublished")
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

func (svc *serviceContext) calculateVisibility(tgtObj uvaeasystore.EasyStoreObject) string {
	fields := tgtObj.Fields()
	visibility := fields["default-visibility"]
	if tgtObj.Namespace() == svc.Namespaces.oa && visibility == "embargo" || tgtObj.Namespace() == svc.Namespaces.etd && visibility == "uva" {
		releaseDateStr := fields["embargo-release"]
		if releaseDateStr == "" {
			// no release date means forever embargoed; just return the default visibility (embargo or uva)
			return visibility
		}

		releaseDate, err := time.Parse(time.RFC3339, releaseDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse embargo release date [%s]: %s", releaseDateStr, err.Error())
			return visibility
		}

		if time.Now().After(releaseDate) {
			return fields["embargo-release-visibility"]
		}
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
	log.Printf("INFO: request to delete etd work %s", workID)
	err := svc.deleteWork(svc.Namespaces.etd, workID)
	if err != nil {
		log.Printf("ERROR: unable to delete etd work %s: %s", workID, err.Error())
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

func (svc *serviceContext) publishWork(namespace string, workID string) error {
	log.Printf("INFO: publish %s work %s", namespace, workID)
	tgtObj, err := svc.EasyStore.GetByKey(namespace, workID, uvaeasystore.BaseComponent|uvaeasystore.Fields)
	if err != nil {
		return fmt.Errorf("unable to get work %s", workID)
	}

	fields := tgtObj.Fields()
	if fields["draft"] == "false" {
		return fmt.Errorf("%s is already published", workID)
	}
	fields["draft"] = "false"
	fields["publish-date"] = time.Now().Format(time.RFC3339)
	_, err = svc.EasyStore.Update(tgtObj, uvaeasystore.Fields)
	if err != nil {
		return fmt.Errorf("publish failed: %s", err.Error())
	}
	svc.publishEvent(uvalibrabus.EventWorkPublish, svc.Namespaces.etd, tgtObj.Id())
	return nil
}

func (svc *serviceContext) unpublishWork(namespace string, workID string) error {
	log.Printf("INFO: get work %s %s for unpublish", namespace, workID)
	tgtObj, err := svc.EasyStore.GetByKey(namespace, workID, uvaeasystore.BaseComponent|uvaeasystore.Fields)
	if err != nil {
		return fmt.Errorf("unable to get work %s", workID)
	}

	fields := tgtObj.Fields()
	if fields["draft"] == "true" {
		return fmt.Errorf("%s is not published", workID)
	}
	fields["draft"] = "true"
	delete(fields, "publish-date")
	_, err = svc.EasyStore.Update(tgtObj, uvaeasystore.Fields)
	if err != nil {
		return fmt.Errorf("unpublish failed: %s", err.Error())
	}
	svc.publishEvent(uvalibrabus.EventWorkUnpublish, svc.Namespaces.etd, tgtObj.Id())
	return nil
}

func (svc *serviceContext) canAccessWork(c *gin.Context, tgtObj uvaeasystore.EasyStoreObject) workAccess {
	// enforce visibility; (admins/authors can always view all)
	//    METADATA: visible to all - except draft content is author/admin only
	//    FILES: open = visible to all; uva = only visible to uva for a limited timeframe; embargo: only author and admin until date
	fields := tgtObj.Fields()
	visibility := svc.calculateVisibility(tgtObj)
	depositor := fields["depositor"]
	isDraft, _ := strconv.ParseBool(fields["draft"])
	resp := workAccess{files: false, metadata: true}
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		if depositor == jwt.ComputeID || jwt.IsAdmin {
			resp.files = true
		}
	} else {
		if isDraft {
			resp.metadata = false
		} else {
			if visibility == "open" {
				resp.files = true
			} else if visibility == "embargo" {
				// embargo work files are only visible to admin / author. that us handled above
				resp.files = false
			} else {
				resp.files = svc.isFromUVA(c)
			}
		}
	}
	return resp
}

func (svc *serviceContext) parseETDWork(tgtObj uvaeasystore.EasyStoreObject, canAccessFiles bool) (*etdWorkDetails, error) {
	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		return nil, fmt.Errorf("unable to read payload: %s", err.Error())
	}
	etdWork, err := librametadata.ETDWorkFromBytes(mdBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse payload: %s", err.Error())
	}

	visibility := svc.calculateVisibility(tgtObj)
	isDraft, _ := strconv.ParseBool(tgtObj.Fields()["draft"])
	resp := etdWorkDetails{
		commonWorkDetails: &commonWorkDetails{
			ID:             tgtObj.Id(),
			IsDraft:        isDraft,
			Version:        tgtObj.VTag(),
			Visibility:     visibility,
			PersistentLink: tgtObj.Fields()["doi"],
			Files:          make([]librametadata.FileData, 0),
		},
		ETDWork: etdWork,
	}
	resp.commonWorkDetails.parseDates(tgtObj)

	if canAccessFiles {
		for _, etdFile := range tgtObj.Files() {
			log.Printf("INFO: add file %s %s to work", etdFile.Name(), etdFile.Url())
			resp.Files = append(resp.Files, librametadata.FileData{Name: etdFile.Name(), MimeType: etdFile.MimeType(), CreatedAt: etdFile.Created()})
		}
	} else {
		log.Printf("INFO: access to files for work %s is restricted", tgtObj.Id())
	}
	return &resp, nil
}

func (svc *serviceContext) parseOAWork(tgtObj uvaeasystore.EasyStoreObject, canAccessFiles bool) (*oaWorkDetails, error) {
	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		return nil, fmt.Errorf("unable to read payload: %s", err.Error())
	}
	oaWork, err := librametadata.OAWorkFromBytes(mdBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse payload: %s", err.Error())
	}

	visibility := svc.calculateVisibility(tgtObj)
	isDraft, _ := strconv.ParseBool(tgtObj.Fields()["draft"])
	resp := oaWorkDetails{
		commonWorkDetails: &commonWorkDetails{
			ID:             tgtObj.Id(),
			IsDraft:        isDraft,
			Version:        tgtObj.VTag(),
			Visibility:     visibility,
			PersistentLink: tgtObj.Fields()["doi"],
			Files:          make([]librametadata.FileData, 0),
		},
		OAWork: oaWork,
	}
	resp.commonWorkDetails.parseDates(tgtObj)

	if canAccessFiles {
		for _, oaFile := range tgtObj.Files() {
			log.Printf("INFO: add file %s %s to work", oaFile.Name(), oaFile.Url())
			resp.Files = append(resp.Files, librametadata.FileData{Name: oaFile.Name(), MimeType: oaFile.MimeType(), CreatedAt: oaFile.Created()})
		}
	} else {
		log.Printf("INFO: access to files for work %s is restricted", tgtObj.Id())
	}

	return &resp, nil
}
