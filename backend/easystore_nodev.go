//go:build !service

package main

import (
	"fmt"

	"github.com/uvalib/easystore/uvaeasystore"
)

// newDevEasyStore is not available in standard builds; see easystore_dev.go
func newDevEasyStore(_ *configData) (uvaeasystore.EasyStore, error) {
	return nil, fmt.Errorf("deveasystoredb requires a build with -tags service")
}
