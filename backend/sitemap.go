package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// WriteRobotsTxt creates or overwrites the robots.txt file.
// To be run once at startup
func (svc *serviceContext) WriteRobotsTxt() {

	robotsTxt := fmt.Appendf(nil, "Sitemap: %s/sitemap.xml", svc.EtdURL)
	robotsPath := filepath.Join("frontend", "public", "robots.txt")

	if err := os.WriteFile(robotsPath, robotsTxt, 0644); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("INFO: %s written [%s]", robotsPath, robotsTxt)
	}
}

// GetSitemap is a handler function that serves the sitemap.xml file.
func (svc *serviceContext) GetSitemap(c *gin.Context) {
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

	payload := map[string]any{"filter": []string{"fields.draft=false"},
		"fields": []string{"id", "modified"},
		"limit":  5000,
	}
	url := fmt.Sprintf("%s/indexes/works/documents/fetch", svc.IndexURL)
	rawResp, respErr := svc.sendPostRequest(url, payload)
	if respErr != nil {
		log.Printf("ERROR: Sitemap search for works failed: %s", respErr.Message)
		return nil, fmt.Errorf(respErr.Message)
	}

	var jsonResp documentResp
	err := json.Unmarshal(rawResp, &jsonResp)
	if err != nil {
		log.Printf("ERROR: unable to parse response: %s", err.Error())
		return nil, err
	}

	// log.Printf("INFO: %+v", jsonResp)

	urls := []sitemapURL{}
	for _, result := range jsonResp.Results {
		url := sitemapURL{
			Loc:     fmt.Sprintf("%s/public/etd/%s", baseURL, result.ID),
			LastMod: result.ModifiedAt,
		}
		urls = append(urls, url)
	}

	urlSet := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/1.1",
		URLs:  urls,
	}

	return &urlSet, nil
}
