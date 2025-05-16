package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

type searchHit struct {
	ID          string                        `json:"id"`
	Title       string                        `json:"title"`
	Author      librametadata.ContributorData `json:"author"`
	Source      string                        `json:"source"`
	Visibility  string                        `json:"visibility"`
	CreatedAt   time.Time                     `json:"createdAt"`
	ModifiedAt  *time.Time                    `json:"modifiedAt,omitempty"`
	PublishedAt *time.Time                    `json:"publishedAt,omitempty"`
}

type searchResp struct {
	Hits []struct {
		ID       string `json:"id"`
		Metadata struct {
			Version string                        `json:"version"`
			Program string                        `json:"program"`
			Degree  string                        `json:"degree"`
			Title   string                        `json:"title"`
			Author  librametadata.ContributorData `json:"author"`
		} `json:"metadata"`
		Fields struct {
			CreateDate        string `json:"create-date"`
			DefaultVisibility string `json:"default-visibility"`
			Depositor         string `json:"depositor"`
			Doi               string `json:"doi"`
			Draft             string `json:"draft"`
			PublishDate       string `json:"publish-date"`
			ModifyDate        string `json:"modify-date"`
			Source            string `json:"source"`
			SourceID          string `json:"source-id"`
		} `json:"fields"`
	} `json:"hits"`
}

func (hit *searchHit) parseDates(esObj uvaeasystore.EasyStoreObject) {
	log.Printf("INFO: parse dates for %s", hit.ID)
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

func (svc *serviceContext) adminSearch(c *gin.Context) {
	qStr := c.Query("q")
	if qStr == "" {
		log.Printf("INFO: missing query")
		c.String(http.StatusBadRequest, "query is required")
		return
	}

	log.Printf("INFO: admin search for works with [%s]", qStr)
	startTime := time.Now()
	payload := map[string]string{"q": qStr}
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: search for [%s] failed: %s", qStr, respErr.Message)
		c.String(respErr.StatusCode, respErr.Message)
		return
	}

	var jsonResp searchResp
	err := json.Unmarshal(rawResp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]searchHit, 0)
	for _, h := range jsonResp.Hits {
		log.Printf("HIT: %+v", h)
		hit := searchHit{
			ID:     h.ID,
			Title:  h.Metadata.Title,
			Author: h.Metadata.Author,
			Source: h.Fields.Source,
		}
		createDate, _ := time.Parse(time.RFC3339, h.Fields.CreateDate)
		hit.CreatedAt = createDate
		if h.Fields.PublishDate != "" {
			pubDate, _ := time.Parse(time.RFC3339, h.Fields.PublishDate)
			hit.PublishedAt = &pubDate
		}
		if h.Fields.ModifyDate != "" {
			modDate, _ := time.Parse(time.RFC3339, h.Fields.ModifyDate)
			hit.ModifiedAt = &modDate
		}
		resp = append(resp, hit)
	}

	elapsedNanoSec := time.Since(startTime)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)
	log.Printf("INFO: received %d search hits. Elapsed Time: %d (ms)", len(resp), elapsedMS)

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
	resp := make([]searchHit, 0)
	obj, err := hits.Next()
	for err == nil {
		var hit *searchHit
		if obj.Namespace() != svc.Namespace {
			continue
		}

		hit, err = svc.parseETDSearchHit(obj)
		if err != nil {
			log.Printf("ERROR: unable to parse search result %s: %s", obj.Id(), err.Error())
			continue
		}

		visibility := svc.calculateVisibility(obj)
		hit.Visibility = visibility

		resp = append(resp, *hit)
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
		ID:     esObj.Id(),
		Title:  etdWork.Title,
		Author: etdWork.Author,
	}
	hit.parseDates(esObj)
	return &hit, nil
}
