package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// getRobotsTxt serves the robots.txt file.
func (svc *serviceContext) getRobotsTxt(c *gin.Context) {
	robotsTxt := fmt.Sprintf("Sitemap: %s/sitemap.xml", svc.EtdURL)
	c.String(http.StatusOK, robotsTxt)
}

// GetSitemap is a handler function that serves the sitemap.xml file.
func (svc *serviceContext) getSitemap(c *gin.Context) {
	log.Printf("INFO: sitemap requested for %s", svc.EtdURL)
	sitemap, err := generateSitemap(svc, svc.EtdURL)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.XML(http.StatusOK, sitemap)
}

// sitemapURL is a single URL entry in the sitemap
type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

// URLSet is the top-level XML element containing an array of URLs
type urlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// Document response from Meilisearch
type documentResp struct {
	Results []struct {
		ID         string `json:"id"`
		ModifiedAt string `json:"modified"`
	} `json:"results"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

func generateSitemap(svc *serviceContext, baseURL string) (*urlSet, error) {
	offset := 0
	limit := 1000
	done := false
	urls := []sitemapURL{}

	log.Printf("INFO: generate sitemap by requesting identifiers in batches of %d", limit)
	for done == false {
		payload := map[string]any{"filter": []string{"fields.draft=false"},
			"fields": []string{"id", "modified"},
			"offset": offset,
			"limit":  limit,
		}
		url := fmt.Sprintf("%s/indexes/works/documents/fetch", svc.IndexURL)
		rawResp, respErr := svc.sendPostRequest(url, payload)
		if respErr != nil {
			log.Printf("ERROR: Sitemap search for works failed: %s", respErr.Message)
			return nil, fmt.Errorf("%s", respErr.Message)
		}

		var jsonResp documentResp
		err := json.Unmarshal(rawResp, &jsonResp)
		if err != nil {
			log.Printf("ERROR: unable to parse response: %s", err.Error())
			return nil, err
		}

		for _, result := range jsonResp.Results {
			url := sitemapURL{
				Loc:     fmt.Sprintf("%s/public_view/%s", baseURL, result.ID),
				LastMod: result.ModifiedAt,
			}
			urls = append(urls, url)
			if len(urls) >= 50000 {
				log.Printf("ERROR: reached 50,000 max item limit for urls in sitemap response; stopping early.")
				done = true
			}
		}

		if done == false {
			if len(urls) == jsonResp.Total {
				log.Printf("INFO: gathered %d urls for sitemap; done", jsonResp.Total)
				done = true
			} else {
				log.Printf("INFO: received %d of %d results; requesting %d more", len(urls), jsonResp.Total, limit)
				offset += limit
			}
		}
	}

	log.Printf("INFO: returning sitemap with %d entries", len(urls))
	urlSet := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	return &urlSet, nil
}
