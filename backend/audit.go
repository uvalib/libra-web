package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

type auditContext struct {
	computeID string
	namespace string
	workID    string
}

func (svc *serviceContext) getAudits(c *gin.Context) {
	workID := c.Param("id")
	tgtObj, eserr := svc.EasyStore.GetByKey(svc.Namespace, workID, uvaeasystore.AllComponents)
	if eserr != nil {
		if strings.Contains(eserr.Error(), "not exist") {
			log.Printf("INFO: work %s was not found", workID)
			c.String(http.StatusNotFound, fmt.Sprintf("%s was not found", workID))
		} else {
			log.Printf("ERROR: unable to get %s work %s for audit request: %s", svc.Namespace, workID, eserr.Error())
			c.String(http.StatusInternalServerError, eserr.Error())
		}
		return
	}

	access := svc.canAccessWork(c, tgtObj)
	if access.metadata == false {
		log.Printf("INFO: access to work %s audit log is forbidden", tgtObj)
		c.String(http.StatusForbidden, "access to %s is not authorized", workID)
		return
	}

	resp, err := svc.sendGetRequest(fmt.Sprintf("%s?namespace=%s&oid=%s", svc.AuditQueryURL, svc.Namespace, workID))
	if err != nil {
		c.String(err.StatusCode, err.Message)
		return
	}

	auditEvents, auditErr := librametadata.AuditsFromBytes(resp)
	if auditErr != nil {
		log.Printf("ERROR: unable to parse audit results: %s", auditErr.Error())
		c.String(http.StatusInternalServerError, auditErr.Error())
		return
	}

	c.JSON(http.StatusOK, auditEvents)
}

func (svc *serviceContext) auditWorkUpdate(computeID string, etdUpdate etdUpdateRequest, origObj uvaeasystore.EasyStoreObject) {
	origWork, err := svc.parseWork(origObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse easystore work %s to generate audit events: %s", origObj.Id(), err.Error())
	}

	// setup an audit context that contains common data needed by all audit logic
	auditCtx := auditContext{
		computeID: computeID,
		namespace: svc.Namespace,
		workID:    origObj.Id(),
	}

	svc.auditVisibiliy(auditCtx, origObj.Fields()["default-visibility"], etdUpdate.Visibility)
	svc.auditFiles(auditCtx, origObj.Files(), etdUpdate.AddFiles, etdUpdate.DelFiles)

	updateVal := reflect.ValueOf(&etdUpdate.Work).Elem()
	origVal := reflect.ValueOf(origWork.ETDWork).Elem()
	svc.auditWork(auditCtx, origVal, updateVal)
}

func (svc *serviceContext) auditVisibiliy(auditCtx auditContext, origVis string, newVis string) {
	if origVis != newVis {
		auditEvt := uvalibrabus.UvaAuditEvent{
			Who:       auditCtx.computeID,
			FieldName: "default-visibility",
			Before:    origVis,
			After:     newVis,
		}
		svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
	}
}

func (svc *serviceContext) auditFiles(auditCtx auditContext, origFiles []uvaeasystore.EasyStoreBlob, added, deleted []string) {
	log.Printf("INFO: audit files; added %v, deleted %v", added, deleted)
	if len(added) == 0 && len(deleted) == 0 {
		return
	}
	orig := make([]string, 0)
	updated := make([]string, 0)
	for _, esBlob := range origFiles {
		fileName := esBlob.Name()
		orig = append(orig, fileName)
		if slices.Contains(deleted, fileName) == false {
			updated = append(updated, fileName)
		}
	}

	updated = append(updated, added...)
	auditEvt := uvalibrabus.UvaAuditEvent{
		Who:       auditCtx.computeID,
		FieldName: "files",
		Before:    strings.Join(orig, ","),
		After:     strings.Join(updated, ","),
	}
	svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
}

func (svc *serviceContext) auditWork(auditCtx auditContext, origVal reflect.Value, updateVal reflect.Value) {
	// use reflection to determine any metadata changes and report them individually
	// there are 4 categories of data in the work metadata to deal with:
	//     1: string
	//     2: ContributorData
	//     3: []string
	//     4: []ContributorData
	for fieldIdx := 0; fieldIdx < origVal.NumField(); fieldIdx++ {
		fieldName := origVal.Type().Field(fieldIdx).Name
		if fieldName == "SchemaVersion" {
			// no audit events on schema version
			continue
		}

		origValue := origVal.Field(fieldIdx)
		newValue := updateVal.FieldByName(fieldName)
		kind := origVal.Field(fieldIdx).Kind()
		switch kind {
		case reflect.Slice:
			svc.auditSliceField(auditCtx, fieldName, origValue, newValue)
		case reflect.String:
			svc.auditStringField(auditCtx, fieldName, origValue, newValue)
		case reflect.Struct:
			// NOTE: right now contributor is the ONLY nested struct. this needs to change if that changes
			svc.auditContributorField(auditCtx, fieldName, origValue, newValue, -1)
		}
	}
}

func (svc *serviceContext) auditStringField(auditCtx auditContext, fieldName string, origValue, newValue reflect.Value) {
	origStr := origValue.String()
	newStr := newValue.String()
	if origStr == newStr {
		return
	}
	auditEvt := uvalibrabus.UvaAuditEvent{
		Who:       auditCtx.computeID,
		FieldName: fieldName,
		Before:    origStr,
		After:     newStr,
	}
	svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
}

func (svc *serviceContext) auditContributorField(auditCtx auditContext, fieldName string, origValue, newValue reflect.Value, contribIdx int) {
	// all data in contributor is string; simple checks are all thats needed
	if origValue.IsValid() == false {
		for fieldIdx := 0; fieldIdx < newValue.NumField(); fieldIdx++ {
			contribField := newValue.Type().Field(fieldIdx).Name
			changeFieldName := fmt.Sprintf("%s.%s", fieldName, contribField)
			if contribIdx > -1 {
				changeFieldName = fmt.Sprintf("%s[%d].%s", fieldName, contribIdx, contribField)
			}
			auditEvt := uvalibrabus.UvaAuditEvent{
				Who:       auditCtx.computeID,
				FieldName: changeFieldName,
				Before:    "",
				After:     newValue.Field(fieldIdx).String(),
			}
			svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
		}
	} else {
		for fieldIdx := 0; fieldIdx < origValue.NumField(); fieldIdx++ {
			contribField := origValue.Type().Field(fieldIdx).Name
			origFieldValue := origValue.Field(fieldIdx).String()
			newFieldValue := ""
			if newValue.IsValid() {
				newFieldValue = newValue.FieldByName(contribField).String()
			}
			if origFieldValue != newFieldValue {
				changeFieldName := fmt.Sprintf("%s.%s", fieldName, contribField)
				if contribIdx > -1 {
					changeFieldName = fmt.Sprintf("%s[%d].%s", fieldName, contribIdx, contribField)
				}
				auditEvt := uvalibrabus.UvaAuditEvent{
					Who:       auditCtx.computeID,
					FieldName: changeFieldName,
					Before:    origFieldValue,
					After:     newFieldValue,
				}
				svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
			}
		}
	}
}

func (svc *serviceContext) auditSliceField(auditCtx auditContext, fieldName string, origValue, newValue reflect.Value) {
	if origValue.Len() == 0 && newValue.Len() == 0 {
		return
	}

	if origValue.Len() > 0 && origValue.Index(0).Kind() == reflect.String ||
		newValue.Len() > 0 && newValue.Index(0).Kind() == reflect.String {
		svc.auditStringSlice(auditCtx, fieldName, origValue, newValue)
	} else {
		// the only other type of slice used is a slice of structs
		svc.auditStructSlice(auditCtx, fieldName, origValue, newValue)
	}
}

func (svc *serviceContext) auditStructSlice(auditCtx auditContext, fieldName string, origValue, newValue reflect.Value) {
	if origValue.Len() == 0 && newValue.Len() == 0 {
		return
	}

	for idx := 0; idx < origValue.Len(); idx++ {
		var newVal reflect.Value
		if idx < newValue.Len() {
			newVal = newValue.Index(idx)
		}
		svc.auditContributorField(auditCtx, fieldName, origValue.Index(idx), newVal, idx)
	}

	if newValue.Len() > origValue.Len() {
		log.Printf("INFO: new array has %d more entries than original", (newValue.Len() - origValue.Len()))
		for idx := origValue.Len(); idx < newValue.Len(); idx++ {
			var emptyVal reflect.Value
			svc.auditContributorField(auditCtx, fieldName, emptyVal, newValue.Index(idx), idx)
		}
	}
}

func (svc *serviceContext) auditStringSlice(auditCtx auditContext, fieldName string, origValue, newValue reflect.Value) {
	var orig []string
	for i := 0; i < origValue.Len(); i++ {
		orig = append(orig, origValue.Index(i).String())
	}
	var upd []string
	for i := 0; i < newValue.Len(); i++ {
		upd = append(upd, newValue.Index(i).String())
	}

	if reflect.DeepEqual(orig, upd) {
		return
	}

	auditEvt := uvalibrabus.UvaAuditEvent{
		Who:       auditCtx.computeID,
		FieldName: fieldName,
		Before:    strings.Join(orig, "|"),
		After:     strings.Join(upd, "|"),
	}
	svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
}

func (svc *serviceContext) publishAuditEvent(nameSpace, workID string, audit uvalibrabus.UvaAuditEvent) {
	auditDetail, _ := audit.Serialize()
	evt := uvalibrabus.UvaBusEvent{
		EventName:  uvalibrabus.EventFieldUpdate,
		Namespace:  nameSpace,
		Identifier: workID,
		Detail:     auditDetail,
	}

	if svc.Events.DevMode {
		log.Printf("INFO: dev mode work %s:%s send audit event data [%s] to bus [%s] with source [%s]", nameSpace, workID, auditDetail, svc.Events.BusName, svc.Events.EventSource)
	} else {
		err := svc.Events.Bus.PublishEvent(&evt)
		if err != nil {
			log.Printf("ERROR: unable to publish audit event %v : %s", evt, err.Error())
		}
	}
}
