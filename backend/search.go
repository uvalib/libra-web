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
	CreatedAt   time.Time                     `json:"created"`
	ModifiedAt  *time.Time                    `json:"modified,omitempty"`
	PublishedAt *time.Time                    `json:"published,omitempty"`
}

type searchResp struct {
	Total  int64       `json:"total"`
	Offset int64       `json:"offset"`
	Limit  int64       `json:"limit"`
	Hits   []searchHit `json:"hits"`
}

type indexResp struct {
	ProcessingTime int64 `json:"processingTimeMs"`
	Total          int64 `json:"estimatedTotalHits"`
	Offset         int64 `json:"offset"`
	Limit          int64 `json:"limit"`
	Hits           []struct {
		ID       string `json:"id"`
		Metadata struct {
			Version string                        `json:"version"`
			Program string                        `json:"program"`
			Degree  string                        `json:"degree"`
			Title   string                        `json:"title"`
			Author  librametadata.ContributorData `json:"author"`
		} `json:"metadata"`
		Fields struct {
			CreateDate               string `json:"create-date"`
			EmbargoRelaeseDate       string `json:"embargo-release"`
			EmbargoRelaeseVisibility string `json:"embargo-release-visibility"`
			DefaultVisibility        string `json:"default-visibility"`
			Depositor                string `json:"depositor"`
			Doi                      string `json:"doi"`
			Draft                    string `json:"draft"`
			PublishDate              string `json:"publish-date"`
			ModifyDate               string `json:"modify-date"`
			Source                   string `json:"source"`
			SourceID                 string `json:"source-id"`
		} `json:"fields"`
	} `json:"hits"`
}

func (svc *serviceContext) parseIndexSearchHits(rawResp indexResp) searchResp {
	resp := searchResp{Total: rawResp.Total, Offset: rawResp.Offset, Limit: rawResp.Limit, Hits: make([]searchHit, 0)}
	for _, h := range rawResp.Hits {

		visibility := svc.calculateVisibility(h.Fields.DefaultVisibility, h.Fields.EmbargoRelaeseDate, h.Fields.EmbargoRelaeseVisibility)
		hit := searchHit{
			ID:         h.ID,
			Title:      h.Metadata.Title,
			Author:     h.Metadata.Author,
			Source:     h.Fields.Source,
			Visibility: visibility,
		}

		hit.CreatedAt = parseDate(h.Fields.CreateDate)
		if h.Fields.PublishDate != "" {
			date := parseDate(h.Fields.PublishDate)
			hit.PublishedAt = &date
		}
		if h.Fields.ModifyDate != "" {
			date := parseDate(h.Fields.ModifyDate)
			hit.ModifiedAt = &date
		}
		resp.Hits = append(resp.Hits, hit)
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
	sort := []string{"metadata.title:asc"}
	payload := map[string]any{"filter": fmt.Sprintf("fields.depositor=%s", computeID), "limit": 250, "sort": sort}
	url := fmt.Sprintf("%s/indexes/works/search", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: search for %s works failed: %s", computeID, respErr.Message)
		c.String(respErr.StatusCode, respErr.Message)
		return
	}

	var jsonResp indexResp
	err := json.Unmarshal(rawResp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := svc.parseIndexSearchHits(jsonResp)
	c.JSON(http.StatusOK, resp)
}
