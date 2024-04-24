package main

import (
	"log"
	"reflect"
	"strings"

	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

type auditContext struct {
	computeID string
	namespace string
	workID    string
	fieldName string
	origValue reflect.Value
	newValue  reflect.Value
}

func (svc *serviceContext) auditOAWorkUpdate(computeID string, oaUpdate oaUpdateRequest, origObj uvaeasystore.EasyStoreObject) {
	origWork, err := svc.parseOAWork(origObj, true)
	if err != nil {
		log.Printf("ERROR: unable to parse easystore oa work %s to generate audit events: %s", origObj.Id(), err.Error())
	}

	// first check for changes in visibility (field instead of metadata)
	if origObj.Fields()["default-visibility"] != oaUpdate.Visibility {
		auditEvt := uvalibrabus.UvaAuditEvent{
			Who:       computeID,
			FieldName: "default-visibility",
			Before:    origObj.Fields()["default-visibility"],
			After:     oaUpdate.Visibility,
		}
		svc.publishAuditEvent(svc.Namespaces.oa, origObj.Id(), auditEvt)
	}

	// next, use reflection to determine any metadata changes and report them individually
	updateVal := reflect.ValueOf(&oaUpdate.Work).Elem()
	origVal := reflect.ValueOf(origWork.OAWork).Elem()
	for fieldIdx := 0; fieldIdx < origVal.NumField(); fieldIdx++ {
		fieldName := origVal.Type().Field(fieldIdx).Name
		if fieldName == "SchemaVersion" {
			// no audit events on schema version
			continue
		}

		auditCtx := auditContext{computeID: computeID,
			namespace: svc.Namespaces.oa,
			workID:    origObj.Id(),
			fieldName: fieldName,
			origValue: origVal.Field(fieldIdx),
			newValue:  updateVal.Field(fieldIdx),
		}
		kind := origVal.Field(fieldIdx).Kind()
		switch kind {
		case reflect.Slice:
			svc.auditSliceField(auditCtx)
		case reflect.String:
			svc.auditStringField(auditCtx)
		}
	}
}

func (svc *serviceContext) auditStringField(auditCtx auditContext) {
	origStr := auditCtx.origValue.String()
	newStr := auditCtx.newValue.String()
	if origStr == newStr {
		return
	}
	auditEvt := uvalibrabus.UvaAuditEvent{
		Who:       auditCtx.computeID,
		FieldName: auditCtx.fieldName,
		Before:    origStr,
		After:     newStr,
	}
	svc.publishAuditEvent(auditCtx.namespace, auditCtx.workID, auditEvt)
}

func (svc *serviceContext) auditSliceField(auditCtx auditContext) {
	if auditCtx.origValue.Len() == 0 && auditCtx.newValue.Len() == 0 {
		return
	}

	if auditCtx.origValue.Len() > 0 && auditCtx.origValue.Index(0).Kind() == reflect.String ||
		auditCtx.newValue.Len() > 0 && auditCtx.newValue.Index(0).Kind() == reflect.String {
		svc.auditStringSlice(auditCtx)
	} else {
		// the only other type of slice used is a slice of structs
		svc.auditStructSlice(auditCtx)
	}
}

func (svc *serviceContext) auditStructSlice(auditCtx auditContext) {
	// TODO
}

func (svc *serviceContext) auditStringSlice(auditCtx auditContext) {
	var orig []string
	for i := 0; i < auditCtx.origValue.Len(); i++ {
		orig = append(orig, auditCtx.origValue.Index(i).String())
	}
	var upd []string
	for i := 0; i < auditCtx.newValue.Len(); i++ {
		upd = append(upd, auditCtx.newValue.Index(i).String())
	}

	if reflect.DeepEqual(orig, upd) {
		return
	}

	auditEvt := uvalibrabus.UvaAuditEvent{
		Who:       auditCtx.computeID,
		FieldName: auditCtx.fieldName,
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
		err := svc.Events.Bus.PublishEvent(evt)
		if err != nil {
			log.Printf("ERROR: unable to publish audit event %v : %s", evt, err.Error())
		}
	}
}
