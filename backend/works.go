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
	ReleaseDate       string `json:"releaseDate"`
	ReleaseVisibility string `json:"releaseVisibility"`
}

type coreWorkDetails struct {
	ID             string                   `json:"id"`
	Version        string                   `json:"version"`
	PersistentLink string                   `json:"persistentLink,omitempty"`
	IsDraft        bool                     `json:"isDraft"`
	Visibility     string                   `json:"visibility"`
	Embargo        *embargoData             `json:"embargo,omitempty"`
	Files          []librametadata.FileData `json:"files"`
	Depositor      string                   `json:"depositor"`
	CreatedAt      time.Time                `json:"createdAt"`
	ModifiedAt     string                   `json:"modifiedAt,omitempty"`
	PublishedAt    string                   `json:"publishedAt,omitempty"`
}

type workDetails struct {
	*coreWorkDetails
	*librametadata.ETDWork
	Source   string `json:"source,omitempty"`
	SourceID string `json:"sourceID,omitempty"`
}

type workAccess struct {
	files    bool
	metadata bool
}

func (svc *serviceContext) getWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get %s work %s", svc.Namespace, workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespace, workID, err.Error())
		if strings.Contains(err.Error(), "not exist") {
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	log.Printf("INFO: check access to work %s", workID)
	access := svc.canAccessWork(c, tgtObj)
	if access.metadata == false {
		log.Printf("INFO: access to %s work %s is forbidden", svc.Namespace, workID)
		c.String(http.StatusForbidden, "access to %s is not authorized", workID)
		return
	}

	etdWork, err := svc.parseWork(tgtObj, access.files)
	if err != nil {
		log.Printf("ERROR: unable to parse work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, etdWork)
}

func (svc *serviceContext) updateWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update work %s", workID)
	var etdReq etdUpdateRequest
	err := c.ShouldBindJSON(&etdReq)
	if err != nil {
		log.Printf("ERROR: bad payload in update request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// load the work - including all current files. Thsi list is necessary for audits of the file list
	log.Printf("INFO: load existing work %s", workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: get work %s for update failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// NOTE: this call has already been thru user or admin middleware, so claims will be present
	claims := getJWTClaims(c)
	if (claims.ComputeID == tgtObj.Fields()["depositor"] || claims.isAdmin()) == false {
		log.Printf("INFO: unauthorized attempt to update work %s", workID)
		c.String(http.StatusForbidden, "you do not have permission to update this work")
		return
	}

	svc.auditWorkUpdate(claims.ComputeID, etdReq, tgtObj)

	// An ETDWork does not serialize the same way as an EasyStoreMetadata object
	// does when being managed by json.Marshal/json.Unmarshal so we wrap it in an object that
	// behaves appropriately
	pl, err := etdReq.Work.Payload()
	if err != nil {
		log.Printf("ERROR: serializing ETDWork: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	tgtObj.SetMetadata(uvaeasystore.NewEasyStoreMetadata(etdReq.Work.MimeType(), pl))

	// update fields
	fields := tgtObj.Fields()
	fields["modify-date"] = time.Now().UTC().Format(svc.TimeFormat)
	fields["default-visibility"] = etdReq.Visibility
	if etdReq.Visibility == "uva" || etdReq.Visibility == "embargo" {
		if etdReq.EmbargoReleaseDate == "" {
			log.Printf("INFO: work %s set for a forever embargo", tgtObj.Id())
			fields["embargo-release"] = ""
			delete(fields, "embargo-release-visibility")
		} else {
			// For non-admin users, visibility must be public within 5 years per provost
			endDate, dateErr := time.Parse(svc.TimeFormat, etdReq.EmbargoReleaseDate)
			if dateErr != nil {
				log.Printf("INFO: reject limited visibiity end date langer than 5 years: %s", etdReq.EmbargoReleaseDate)
				c.String(http.StatusBadRequest, fmt.Sprintf("invalid end date: %s", etdReq.EmbargoReleaseDate))
				return
			}
			if etdReq.Visibility == "uva" && claims.isAdmin() == false {
				maxDate := time.Now().AddDate(5, 0, 0) // now + 5 years
				if endDate.After(maxDate) {
					log.Printf("INFO: reject limited visibiity end date langer than 5 years: %s", etdReq.EmbargoReleaseDate)
					c.String(http.StatusBadRequest, "limited visibilty end date must be less than five years from today")
					return
				}
			}
			fields["embargo-release"] = endDate.UTC().Format(svc.TimeFormat)
			fields["embargo-release-visibility"] = etdReq.EmbargoReleaseVisibility
		}
	} else {
		delete(fields, "embargo-release")
		delete(fields, "embargo-release-visibility")
	}
	tgtObj.SetFields(fields)

	_, err = svc.EasyStore.ObjectUpdate(tgtObj, uvaeasystore.Fields|uvaeasystore.Metadata)
	if err != nil {
		log.Printf("ERROR: unable to update work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// update files if necessary
	if len(etdReq.AddFiles) != 0 || len(etdReq.DelFiles) != 0 {
		err := svc.updateWorkFiles(tgtObj, etdReq.AddFiles, etdReq.DelFiles)
		if err != nil {
			log.Printf("ERROR: unable to update work files with new add [%v] and delete [%v] %s", etdReq.AddFiles, etdReq.DelFiles, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// reload work to get latest files and vtag
	log.Printf("INFO: get %s work %s", svc.Namespace, workID)
	updatedObj, err := svc.EasyStore.ObjectGetByKey(tgtObj.Namespace(), tgtObj.Id(), uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get updated work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	resp, _ := svc.parseWork(updatedObj, true)
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) updateWorkFiles(esObj uvaeasystore.EasyStoreObject, addFiles, delFiles []string) error {
	for _, delFile := range delFiles {
		log.Printf("INFO: delete %s from work %s", delFile, esObj.Id())
		err := svc.EasyStore.FileDelete(esObj.Namespace(), esObj.Id(), delFile)
		if err != nil {
			return fmt.Errorf("unable to delete %s: %s", delFile, err.Error())
		}
	}

	uploadDir := path.Join("/tmp", esObj.Id())
	for _, addFile := range addFiles {
		fullPath := path.Join(uploadDir, addFile)
		log.Printf("INFO: add %s to work %s", fullPath, esObj.Id())
		fileBytes, fileErr := os.ReadFile(fullPath)
		if fileErr != nil {
			return fmt.Errorf("unable to read %s: %s", fullPath, fileErr.Error())
		}
		mimeType := http.DetectContentType(fileBytes)
		log.Printf("INFO: create easystore file blob for %s with size %d and mime type %s", addFile, len(fileBytes), mimeType)
		esBlob := uvaeasystore.NewEasyStoreBlob(addFile, mimeType, fileBytes)
		err := svc.EasyStore.FileCreate(esObj.Namespace(), esObj.Id(), esBlob)
		if err != nil {
			return fmt.Errorf("unable to add %s: %s", addFile, err.Error())
		}
	}
	return nil
}

func (svc *serviceContext) publishWork(c *gin.Context) {
	workID := c.Param("id")

	// NOTE: this call has already been thru user or admin middleware, so claims will be present
	claims := getJWTClaims(c)

	log.Printf("INFO: %s requests work %s publish", claims.ComputeID, workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.BaseComponent|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: unable to get work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if (claims.ComputeID == tgtObj.Fields()["depositor"] || claims.isAdmin()) == false {
		log.Printf("INFO: unauthorized attempt by %s to publish work %s", claims.ComputeID, workID)
		c.String(http.StatusForbidden, "you do not have permission to publish this work")
		return
	}

	fields := tgtObj.Fields()
	if fields["draft"] == "false" {
		log.Printf("INFO: %s is already published", workID)
		c.String(http.StatusConflict, fmt.Sprintf("%s is already published", workID))
		return
	}

	svc.auditPublicationChange(claims.ComputeID, tgtObj, true)

	fields["draft"] = "false"
	fields["publish-date"] = time.Now().UTC().Format(svc.TimeFormat)
	_, err = svc.EasyStore.ObjectUpdate(tgtObj, uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: publish %s failed: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("publish failed: %s", err.Error()))
		return
	}
	svc.publishEvent(uvalibrabus.EventWorkPublish, svc.Namespace, tgtObj.Id())
	c.String(http.StatusOK, "published")
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

func (svc *serviceContext) calculateVisibility(defaultVisibility string, releaseDateStr string, releaseVisibility string) string {
	// ETD can have an embaro period when files can only be accessed by admin/author
	// ETD works can have uva visibility for a limited time
	// In either case, the date is held in embargo-release and the visibility in embargo-release-visibility
	if defaultVisibility == "embargo" || defaultVisibility == "uva" {
		if releaseDateStr == "" {
			// no release date means forever embargoed; just return the default visibility (embargo or uva)
			return defaultVisibility
		}

		releaseDate, err := time.Parse(svc.TimeFormat, releaseDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse embargo release date [%s]: %s", releaseDateStr, err.Error())
			return defaultVisibility
		}

		if time.Now().After(releaseDate) {
			return releaseVisibility
		}
	}
	return defaultVisibility
}

func (svc *serviceContext) renameFile(c *gin.Context) {
	workID := c.Param("id")
	var renameReq struct {
		OriginalName string `json:"orignalName"`
		NewName      string `json:"newName"`
	}
	err := c.ShouldBindJSON(&renameReq)
	if err != nil {
		log.Printf("ERROR: bad payload in rename request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("INFO: request to rename file %s from work %s to %s", renameReq.OriginalName, workID, renameReq.NewName)

	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s for file rename: %s", svc.Namespace, workID, err.Error())
		if strings.Contains(err.Error(), "not exist") {
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	rnErr := svc.EasyStore.FileRename(svc.Namespace, tgtObj.Id(), renameReq.OriginalName, renameReq.NewName)
	if rnErr != nil {
		log.Printf("ERROR: rename %s to %s failed: %s", renameReq.OriginalName, renameReq.NewName, rnErr.Error())
		c.String(http.StatusInternalServerError, rnErr.Error())
		return
	}

	c.String(http.StatusOK, renameReq.NewName)
}

func (svc *serviceContext) downloadFile(c *gin.Context) {
	workID := c.Param("id")
	tgtFile := c.Param("name")
	log.Printf("INFO: request to download file %s from work %s", tgtFile, workID)
	tgtObj, err := svc.EasyStore.ObjectGetByKey(svc.Namespace, workID, uvaeasystore.Files)
	if err != nil {
		log.Printf("ERROR: get %s work %s for download failed: %s", svc.Namespace, workID, err.Error())
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
		log.Printf("INFO: file %s not found in %s work %s", tgtFile, svc.Namespace, workID)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", tgtFile))
		return
	}

	log.Printf("INFO: %s download link %s", tgtFile, dlFile.Url())
	c.String(http.StatusOK, dlFile.Url())
}

func (svc *serviceContext) canAccessWork(c *gin.Context, tgtObj uvaeasystore.EasyStoreObject) workAccess {
	// enforce visibility; (admins/authors can always view all)
	//    METADATA: visible to all - except draft content is author/admin only
	//    FILES: open = visible to all; uva = only visible to uva for a limited timeframe; embargo: only author and admin until date
	fields := tgtObj.Fields()
	visibility := svc.calculateVisibility(fields["default-visibility"], fields["embargo-release"], fields["embargo-release"])
	depositor := fields["depositor"]
	isDraft, _ := strconv.ParseBool(fields["draft"])
	log.Printf("INFO: check if work %s with visibility %s can be accessed", tgtObj.Id(), visibility)
	resp := workAccess{files: false, metadata: true}
	if isSignedIn(c) {
		jwt := getJWTClaims(c)
		log.Printf("INFO: work %s accessed by signed in user %s", tgtObj.Id(), jwt.ComputeID)
		if depositor == jwt.ComputeID || jwt.isAdmin() {
			log.Printf("INFO: user %s is admin or author of work %s and has full access", jwt.ComputeID, tgtObj.Id())
			resp.files = true
		}
	} else {
		if isDraft {
			log.Printf("INFO: work %s is a draft cannot be accessed", tgtObj.Id())
			resp.metadata = false
		} else {
			switch visibility {
			case "open":
				log.Printf("INFO: work %s is public and is fully visibile", tgtObj.Id())
				resp.files = true
			case "embargo":
				// embargo work files are only visible to admin / author. that us handled above
				log.Printf("INFO: work %s is embargoed and only metadata is visible", tgtObj.Id())
				resp.files = false
			default:
				resp.files = svc.isFromUVA(c)
				log.Printf("INFO: work %s is limited to uva users; user uva status %t", tgtObj.Id(), resp.files)
			}
		}
	}
	return resp
}

func (svc *serviceContext) parseWork(tgtObj uvaeasystore.EasyStoreObject, canAccessFiles bool) (*workDetails, error) {
	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		return nil, fmt.Errorf("unable to read payload: %s", err.Error())
	}
	etdWork, err := librametadata.ETDWorkFromBytes(mdBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse payload: %s", err.Error())
	}

	fields := tgtObj.Fields()
	visibility := svc.calculateVisibility(fields["default-visibility"], fields["embargo-release"], fields["embargo-release-visibility"])
	isDraft, _ := strconv.ParseBool(fields["draft"])
	resp := workDetails{
		coreWorkDetails: &coreWorkDetails{
			ID:             tgtObj.Id(),
			CreatedAt:      tgtObj.Created(),
			IsDraft:        isDraft,
			Version:        tgtObj.VTag(),
			Visibility:     visibility,
			Depositor:      tgtObj.Fields()["depositor"],
			PersistentLink: tgtObj.Fields()["doi"],
			Files:          make([]librametadata.FileData, 0),
		},
		ETDWork: etdWork,
	}

	if resp.IsDraft == false {
		resp.PublishedAt = tgtObj.Fields()["publish-date"]
	}
	modDateStr, modified := tgtObj.Fields()["modify-date"]
	if modified {
		resp.ModifiedAt = modDateStr
	}

	embargoDateStr, exist := tgtObj.Fields()["embargo-release"]
	if exist {
		embInfo := embargoData{ReleaseDate: embargoDateStr, ReleaseVisibility: tgtObj.Fields()["embargo-release-visibility"]}
		resp.Embargo = &embInfo
	}

	resp.Source = tgtObj.Fields()["source"]
	resp.SourceID = tgtObj.Fields()["source-id"]

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
