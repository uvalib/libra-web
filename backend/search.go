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
	ID            string     `json:"id"`
	Namespace     string     `json:"namespace"`
	Title         string     `json:"title"`
	DateCreated   time.Time  `json:"dateCreated"`
	DatePublished *time.Time `json:"datePublished,omitempty"`
	Visibility    string     `json:"visibility"`
}

func (svc *serviceContext) searchWorks(c *gin.Context) {
	// NOTES: search accepts query params for search:
	//    type=oa|etd, cid=compute_id, title=title
	workType := c.Query("type")
	computeID := c.Query("cid")
	namespace := ""
	if workType == "oa" {
		namespace = svc.Namespaces.oa
	} else if workType == "etd" {
		namespace = svc.Namespaces.etd
	}

	fields := uvaeasystore.DefaultEasyStoreFields()
	if computeID != "" {
		fields["depositor"] = computeID
	}

	log.Printf("INFO: find %s works with fields %v", namespace, fields)
	hits, err := svc.EasyStore.GetByFields(namespace, fields, uvaeasystore.Metadata|uvaeasystore.Fields)
	if err != nil {
		log.Printf("ERROR: search failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: %d hits returned; parsing results", hits.Count())
	resp := make([]*searchHit, 0)
	obj, err := hits.Next()
	for err == nil {
		var hit *searchHit
		if obj.Namespace() == svc.Namespaces.oa {
			hit, err = svc.parseOASearchHit(obj)
			if err != nil {
				log.Printf("ERROR: unable to parse oa search result %s: %s", obj.Id(), err.Error())
				continue
			}
		} else if obj.Namespace() == svc.Namespaces.etd {
			hit, err = svc.parseETDSearchHit(obj)
			if err != nil {
				log.Printf("ERROR: unable to parse etd search result %s: %s", obj.Id(), err.Error())
				continue
			}
		} else {
			log.Printf("ERROR: search hit %s has unsupported namespace %s; skipping", obj.Id(), obj.Namespace())
			continue
		}

		resp = append(resp, hit)
		obj, err = hits.Next()
	}
	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) parseOASearchHit(esObj uvaeasystore.EasyStoreObject) (*searchHit, error) {
	oaWorkBytes, objErr := esObj.Metadata().Payload()
	if objErr != nil {
		return nil, fmt.Errorf("unable to get object %s payload: %s", esObj.Id(), objErr.Error())
	}

	oaWork, objErr := librametadata.OAWorkFromBytes(oaWorkBytes)
	if objErr != nil {
		return nil, objErr
	}
	visibility := esObj.Fields()["default-visibility"]
	if visibility == "embargo" {
		visibility = calculateVisibility(esObj.Fields())
	}

	hit := searchHit{
		ID:          esObj.Id(),
		Namespace:   svc.Namespaces.oa,
		Title:       oaWork.Title,
		Visibility:  visibility,
		DateCreated: esObj.Created(),
	}

	pubDateStr, published := esObj.Fields()["publish-date"]
	if published {
		pubDate, err := time.Parse(time.RFC3339, pubDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse publish-date [%s]: %s", pubDateStr, err.Error())
			pubDate = time.Now()
		}
		hit.DatePublished = &pubDate
	}
	return &hit, nil
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
		ID:          esObj.Id(),
		Namespace:   svc.Namespaces.oa,
		Title:       etdWork.Title,
		Visibility:  esObj.Fields()["default-visibility"],
		DateCreated: esObj.Created(),
	}
	pubDateStr, published := esObj.Fields()["publish-date"]
	if published {
		pubDate, err := time.Parse(time.RFC3339, pubDateStr)
		if err != nil {
			log.Printf("ERROR: unable to parse publish-date [%s]: %s", pubDateStr, err.Error())
			pubDate = time.Now()
		}
		hit.DatePublished = &pubDate
	}
	return &hit, nil
}
