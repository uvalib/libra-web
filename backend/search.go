package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

	// log.Printf("INDEX RESP: %s", rawResp)

	var jsonResp searchResp
	err := json.Unmarshal(rawResp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := parseIndexSearchHits(jsonResp)

	elapsedNanoSec := time.Since(startTime)
	elapsedMS := int64(elapsedNanoSec / time.Millisecond)
	log.Printf("INFO: received %d search hits. Elapsed Time: %d (ms)", len(resp), elapsedMS)

	c.JSON(http.StatusOK, resp)
}

func parseIndexSearchHits(indexResp searchResp) []searchHit {
	resp := make([]searchHit, 0)
	for _, h := range indexResp.Hits {

		// TODO:
		// visibility := svc.calculateVisibility(obj)
		hit := searchHit{
			ID:         h.ID,
			Title:      h.Metadata.Title,
			Author:     h.Metadata.Author,
			Source:     h.Fields.Source,
			Visibility: h.Fields.DefaultVisibility,
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
	return resp
}

func (svc *serviceContext) userSearch(c *gin.Context) {
	computeID := c.Query("cid")
	if computeID == "" {
		log.Printf("INFO: invalid search for user works without a compute id")
		c.String(http.StatusBadRequest, "cid is required")
		return
	}

	log.Printf("INFO: find user %s works", computeID)
	payload := map[string]string{"filter": fmt.Sprintf("fields.depositor=%s", computeID)}
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: search for %s works failed: %s", computeID, respErr.Message)
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

	resp := parseIndexSearchHits(jsonResp)
	c.JSON(http.StatusOK, resp)
}
