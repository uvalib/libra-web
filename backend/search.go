package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
)

func (svc *serviceContext) searchWorks(c *gin.Context) {
	// NOTES: search accepts query params for search:
	//    type=oa|etd, cid=compute_id, title=title
	workType := c.Query("type")
	computeID := c.Query("cid")
	tgtTitle := c.Query("title")
	if workType != "oa" && workType != "etd" && workType != "" {
		log.Printf("INFO: invalid type [%s] specified", workType)
		c.String(http.StatusBadRequest, fmt.Sprintf("'%s' is not a valid type", workType))
		return
	}

	fields := uvaeasystore.DefaultEasyStoreFields()
	if computeID != "" {
		fields["depositor"] = computeID
	}
	if tgtTitle != "" {
		fields["title"] = tgtTitle
	}
	log.Printf("INFO: find %s works with fields %v", workType, fields)
	hits, err := svc.EasyStore.GetByFields(workType, fields, uvaeasystore.Metadata)
	if err != nil {
		log.Printf("ERROR: search failed: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: %d hits returned", hits.Count())
	c.JSON(http.StatusOK, hits)
}
