//go:build service

package main

import (
	"log"
	"os"

	"github.com/uvalib/easystore/uvaeasystore"
)

// newDevEasyStore creates an easystore that talks directly to a postgres DB and S3 bucket instead of the
// easystore proxy. Dev use only; the direct implementation is only available when built with -tags service.
// S3 endpoint, region and credentials come from the standard AWS_* environment variables.
func newDevEasyStore(cfg *configData) (uvaeasystore.EasyStore, error) {
	return uvaeasystore.NewEasyStore(uvaeasystore.DatastoreS3Config{
		Bucket:              cfg.dev.esBucket,
		SignerAccessKey:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SignerSecretKey:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SignerExpireMinutes: 60,
		DbHost:              cfg.db.host,
		DbPort:              cfg.db.port,
		DbName:              cfg.dev.esDB,
		DbUser:              cfg.db.user,
		DbPassword:          cfg.db.pass,
		DbTimeout:           15,
		Log:                 log.Default(),
	})
}
