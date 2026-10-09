//go:build service

// seed creates sample works in the local easystore DB and (re)builds the local search index
// from everything in easystore. Dev use only; see dev/README.md
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

const timeFormat = "2006-01-02T15:04:05Z"

var indexURL string

func main() {
	var dbHost, dbName, dbUser, dbPass, bucket, namespace, cid, first, last string
	var dbPort int
	var create bool
	flag.StringVar(&dbHost, "dbhost", "localhost", "easystore DB host")
	flag.IntVar(&dbPort, "dbport", 55432, "easystore DB port")
	flag.StringVar(&dbName, "dbname", "easystore", "easystore DB name")
	flag.StringVar(&dbUser, "dbuser", "libraweb", "easystore DB user")
	flag.StringVar(&dbPass, "dbpass", "pass", "easystore DB password")
	flag.StringVar(&bucket, "bucket", "easystore-local", "easystore S3 bucket")
	flag.StringVar(&indexURL, "index", "http://localhost:7700", "search index URL")
	flag.StringVar(&namespace, "namespace", "libraetd", "easystore namespace")
	flag.BoolVar(&create, "create", true, "create a sample draft work before reindexing")
	flag.StringVar(&cid, "cid", "dev1x", "compute ID of the sample work author / depositor")
	flag.StringVar(&first, "first", "Local", "sample work author first name")
	flag.StringVar(&last, "last", "Developer", "sample work author last name")
	flag.Parse()

	// S3 endpoint, region and credentials come from the standard AWS_* environment variables
	es, err := uvaeasystore.NewEasyStore(uvaeasystore.DatastoreS3Config{
		Bucket: bucket, SignerAccessKey: os.Getenv("AWS_ACCESS_KEY_ID"), SignerSecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SignerExpireMinutes: 60, DbHost: dbHost, DbPort: dbPort, DbName: dbName, DbUser: dbUser, DbPassword: dbPass,
		DbTimeout: 15, Log: log.Default(),
	})
	if err != nil {
		log.Fatalf("unable to connect to easystore: %s", err.Error())
	}
	defer es.Close()

	if create {
		id, err := createSampleWork(es, namespace, cid, first, last)
		if err != nil {
			log.Fatalf("unable to create sample work: %s", err.Error())
		}
		log.Printf("created sample work %s for %s", id, cid)
		log.Printf("edit it at http://localhost:8080/etd/%s", id)
	}

	if err := setupIndex(); err != nil {
		log.Fatalf("unable to set up index: %s", err.Error())
	}
	if err := reindex(es, namespace); err != nil {
		log.Fatalf("reindex failed: %s", err.Error())
	}
}

func createSampleWork(es uvaeasystore.EasyStore, namespace, cid, first, last string) (string, error) {
	author := librametadata.ContributorData{ComputeID: cid, FirstName: first, LastName: last, Institution: "University of Virginia"}
	work := librametadata.ETDWork{
		Program: "Computer Science",
		Degree:  "PHD (Doctor of Philosophy)",
		Title:   fmt.Sprintf("Sample Thesis %s", time.Now().Format("2006-01-02 15:04:05")),
		Author:  author,
	}
	pl, err := work.Payload()
	if err != nil {
		return "", err
	}

	obj := uvaeasystore.NewEasyStoreObject(namespace, "")
	fields := uvaeasystore.DefaultEasyStoreFields()
	fields["create-date"] = time.Now().UTC().Format(timeFormat)
	fields["draft"] = "true"
	fields["default-visibility"] = ""
	fields["depositor"] = cid
	fields["source"] = "optional"
	obj.SetFields(fields)
	obj.SetMetadata(uvaeasystore.NewEasyStoreMetadata(work.MimeType(), pl))
	created, err := es.ObjectCreate(obj)
	if err != nil {
		return "", err
	}
	return created.Id(), nil
}

// setupIndex creates the works index with the filter/sort settings the backend searches rely on
func setupIndex() error {
	if _, err := meili("POST", "/indexes", map[string]string{"uid": "works", "primaryKey": "id"}); err != nil {
		return err
	}
	_, err := meili("PATCH", "/indexes/works/settings", map[string]any{
		"filterableAttributes": []string{"fields.depositor", "fields.source", "fields.draft",
			"createdUnix", "publishedUnix", "modifiedUnix"},
		"sortableAttributes": []string{"metadata.title", "fields.create-date", "fields.publish-date", "modified"},
	})
	return err
}

func reindex(es uvaeasystore.EasyStore, namespace string) error {
	set, err := es.ObjectGetByFields(namespace, uvaeasystore.DefaultEasyStoreFields(), uvaeasystore.Fields|uvaeasystore.Metadata)
	if err != nil {
		return err
	}
	docs := make([]map[string]any, 0)
	for i := uint(0); i < set.Count(); i++ {
		obj, err := set.Next()
		if err != nil {
			return err
		}
		md := map[string]any{}
		if obj.Metadata() != nil {
			pl, _ := obj.Metadata().Payload()
			json.Unmarshal(pl, &md)
		}
		fields := obj.Fields()
		doc := map[string]any{"id": obj.Id(), "metadata": md, "fields": fields}
		for src, dst := range map[string]string{"create-date": "created", "modify-date": "modified", "publish-date": "published"} {
			if t, err := time.Parse(timeFormat, fields[src]); err == nil {
				doc[dst] = fields[src]
				doc[dst+"Unix"] = t.Unix()
			}
		}
		docs = append(docs, doc)
	}
	if len(docs) == 0 {
		log.Printf("no works to index")
		return nil
	}
	if _, err := meili("POST", "/indexes/works/documents", docs); err != nil {
		return err
	}
	log.Printf("indexed %d works", len(docs))
	return nil
}

// meili sends a request to the search index and waits for the resulting async task to finish
func meili(method, path string, payload any) ([]byte, error) {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(method, indexURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, out)
	}

	var task struct {
		TaskUID *int `json:"taskUid"`
	}
	json.Unmarshal(out, &task)
	if task.TaskUID == nil {
		return out, nil
	}
	for range 100 {
		r, err := http.Get(fmt.Sprintf("%s/tasks/%d", indexURL, *task.TaskUID))
		if err != nil {
			return nil, err
		}
		var status struct {
			Status string `json:"status"`
			Error  *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(r.Body).Decode(&status)
		r.Body.Close()
		switch status.Status {
		case "succeeded":
			return out, nil
		case "failed":
			if status.Error != nil && status.Error.Code == "index_already_exists" {
				return out, nil
			}
			return nil, fmt.Errorf("%s %s task failed: %+v", method, path, status.Error)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("%s %s: timed out waiting for task", method, path)
}
