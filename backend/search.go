package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

type searchHit struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	ComputeID   string     `json:"computeID"`
	CreatedAt   time.Time  `json:"createdAt"`
	ModifiedAt  *time.Time `json:"modifiedAt,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

func (hit *searchHit) parseDates(esObj uvaeasystore.EasyStoreObject) {
	log.Printf("INFO: parse dates for %s", hit.ID)
	hit.CreatedAt = esObj.Created()
	pubDateStr, published := esObj.Fields()["publish-date"]
	if published {
		log.Printf("INFO: has published date %s", pubDateStr)
		pubDate, err := time.Parse(time.RFC3339, pubDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse publish-date [%s]: %s", pubDateStr, err.Error())
			pubDate = time.Now()
		}
		hit.PublishedAt = &pubDate
	}
	modDateStr, modified := esObj.Fields()["modify-date"]
	if modified {
		log.Printf("INFO: has modified date %s", modDateStr)
		modDate, err := time.Parse(time.RFC3339, modDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse modify-date [%s]: %s", modDateStr, err.Error())
			modDate = time.Now()
		}
		hit.ModifiedAt = &modDate
	}
}

type adminSearchHit struct {
	*searchHit
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
}

type userSearchHit struct {
	*searchHit
	Visibility string `json:"visibility"`
}

func (svc *serviceContext) adminSearch(c *gin.Context) {
	computeID := c.Query("cid")
	if computeID == "" {
		log.Printf("INFO: invalid search; missing cid")
		c.String(http.StatusBadRequest, "cid is required")
		return
	}

	log.Printf("INFO: admin search for works by %s", computeID)
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = computeID
	hits, err := svc.EasyStore.GetByFields("", fields, uvaeasystore.Metadata|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: search failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: %d hits returned; parsing results", hits.Count())
	resp := make([]adminSearchHit, 0)
	obj, err := hits.Next()
	for err == nil {
		adminHit := adminSearchHit{Source: obj.Fields()["source"], Namespace: obj.Namespace()}
		var hit *searchHit
		if obj.Namespace() != svc.Namespace {
			log.Printf("INFO: skipping hit with mismatched namespace %s vs expected %s", obj.Namespace(), svc.Namespace)
			continue
		}

		hit, err = svc.parseETDSearchHit(obj)
		if err != nil {
			log.Printf("ERROR: unable to parse search result %s: %s", obj.Id(), err.Error())
			continue
		}

		adminHit.searchHit = hit
		resp = append(resp, adminHit)
		obj, err = hits.Next()
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) userSearch(c *gin.Context) {
	computeID := c.Query("cid")
	if computeID == "" {
		log.Printf("INFO: invalid search for user works without a compute id")
		c.String(http.StatusBadRequest, "cid is required")
		return
	}

	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["depositor"] = computeID

	log.Printf("INFO: find user %s %s works", computeID, svc.Namespace)
	hits, err := svc.EasyStore.GetByFields(svc.Namespace, fields, uvaeasystore.Metadata|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: search failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: %d hits returned; parsing results", hits.Count())
	resp := make([]userSearchHit, 0)
	obj, err := hits.Next()
	for err == nil {
		var hit *searchHit
		if obj.Namespace() != svc.Namespace {
			log.Printf("INFO: skipping hit with mismatched namespace %s vs expected %s", obj.Namespace(), svc.Namespace)
			continue
		}

		hit, err = svc.parseETDSearchHit(obj)
		if err != nil {
			log.Printf("ERROR: unable to parse search result %s: %s", obj.Id(), err.Error())
			continue
		}

		userHit := userSearchHit{searchHit: hit}
		visibility := svc.calculateVisibility(obj)
		userHit.Visibility = visibility

		resp = append(resp, userHit)
		obj, err = hits.Next()
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) parseETDSearchHit(esObj uvaeasystore.EasyStoreObject) (*searchHit, error) {
	etdWorkBytes, objErr := esObj.Metadata().Payload()
	if objErr != nil {
		return nil, fmt.Errorf("unable to get object %s payload: %s", esObj.Id(), objErr.Error())
	}

	etdWork, objErr := librametadata.ETDWorkFromBytes(etdWorkBytes)
	if objErr != nil {
		return nil, objErr
	}
	hit := searchHit{
		ID:        esObj.Id(),
		Title:     etdWork.Title,
		ComputeID: etdWork.Author.ComputeID,
	}
	hit.parseDates(esObj)
	return &hit, nil
}
