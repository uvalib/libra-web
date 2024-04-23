package main

import (
	"log"
	"reflect"

	"github.com/uvalib/easystore/uvaeasystore"
	"github.com/uvalib/librabus-sdk/uvalibrabus"
)

func (svc *serviceContext) auditOAWorkUpdate(computeID string, oaUpdate oaUpdateRequest, original uvaeasystore.EasyStoreObject) {
	origWork, err := svc.parseOAWork(original, true)
	if err != nil {
		log.Printf("ERROR: unable to parse original work %s to generate audit events: %s", original.Id(), err.Error())
	}

	updateVal := reflect.ValueOf(&oaUpdate.Work).Elem()
	origVal := reflect.ValueOf(origWork.OAWork).Elem()
	for fieldIdx := 0; fieldIdx < origVal.NumField(); fieldIdx++ {
		fieldName := origVal.Type().Field(fieldIdx).Name
		origFieldVal := origVal.Field(fieldIdx).String()
		updateFieldVal := updateVal.Field(fieldIdx).String()
		if origFieldVal != updateFieldVal {
			audit := uvalibrabus.UvaAuditEvent{
				Who:       computeID,
				FieldName: fieldName,
				Before:    origFieldVal,
				After:     updateFieldVal,
			}
			auditDetail, _ := audit.Serialize()
			if svc.Events.DevMode {
				log.Printf("INFO: dev mode send audit event %v to bus [%s] with source [%s]", audit, svc.Events.BusName, svc.Events.EventSource)
			} else {
				ev := uvalibrabus.UvaBusEvent{
					EventName:  uvalibrabus.EventFieldUpdate,
					Namespace:  svc.Namespaces.oa,
					Identifier: original.Id(),
					Detail:     auditDetail,
				}
				err := svc.Events.Bus.PublishEvent(ev)
				if err != nil {
					log.Printf("ERROR: unable to publish event %s %s - %s: %s", ev.EventName, ev.Namespace, ev.Identifier, err.Error())
				}
			}
		}
	}
}
