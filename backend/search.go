package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

type searchHit struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	Title       string    `json:"title"`
	DateCreated time.Time `json:"dateCreated"`
	Visibility  string    `json:"visibility"`
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
		objBytes, objErr := obj.Metadata().Payload()
		if objErr != nil {
			log.Printf("ERROR: unable to get object %s paylpoad: %s", obj.Id(), objErr.Error())
			break
		}

		var hit *searchHit
		if obj.Namespace() == svc.Namespaces.oa {
			hit, err = svc.parseOASearchHit(obj.Id(), objBytes, obj.Fields())
			if err != nil {
				log.Printf("ERROR: unable to parse oa search result %s: %s", obj.Id(), err.Error())
				continue
			}
		} else if obj.Namespace() == svc.Namespaces.etd {
			hit, err = svc.parseETDSearchHit(obj.Id(), objBytes, obj.Fields())
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

func (svc *serviceContext) parseOASearchHit(id string, oaWorkBytes []byte, fields uvaeasystore.EasyStoreObjectFields) (*searchHit, error) {
	oaWork, objErr := librametadata.OAWorkFromBytes(oaWorkBytes)
	if objErr != nil {
		return nil, objErr
	}
	visibility := fields["default-visibility"]
	if visibility == "embargo" {
		visibility = calculateVisibility(fields)
	}

	hit := searchHit{ID: id,
		Namespace:   svc.Namespaces.oa,
		Title:       oaWork.Title,
		Visibility:  visibility,
		DateCreated: oaWork.Created(),
	}
	return &hit, nil
}

func (svc *serviceContext) parseETDSearchHit(id string, etdWorkBytes []byte, fields uvaeasystore.EasyStoreObjectFields) (*searchHit, error) {
	etdWork, objErr := librametadata.ETDWorkFromBytes(etdWorkBytes)
	if objErr != nil {
		return nil, objErr
	}
	hit := searchHit{ID: id,
		Namespace:   svc.Namespaces.oa,
		Title:       etdWork.Title,
		Visibility:  fields["default-visibility"],
		DateCreated: etdWork.Created(),
	}
	return &hit, nil
}
