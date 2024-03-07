package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uvalib/easystore/uvaeasystore"
	librametadata "github.com/uvalib/libra-metadata"
)

func (svc *serviceContext) getOAWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: get oa work %s", workID)
	tgtObj, err := svc.EasyStore.GetByKey(svc.Namespaces.oa, workID, uvaeasystore.AllComponents)
	if err != nil {
		log.Printf("ERROR: unable to get %s work %s: %s", svc.Namespaces.oa, workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	mdBytes, err := tgtObj.Metadata().Payload()
	if err != nil {
		log.Printf("ERROR: unable to get metadata paload from respose: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	parsedOAWork, err := librametadata.OAWorkFromBytes(mdBytes)
	if err != nil {
		log.Printf("ERROR: unable to process paypad from work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	resp := versionedOA{ID: tgtObj.Id(), Version: tgtObj.VTag(), OAWork: parsedOAWork, CreatedAt: tgtObj.Created(), ModifiedAt: tgtObj.Modified()}
	c.JSON(http.StatusOK, resp)

}

func (svc *serviceContext) deleteOAWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete oa work %s", workID)
	err := svc.deleteWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: unable to delete oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) deleteETDWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete oa work %s", workID)
	err := svc.deleteWork(svc.Namespaces.oa, workID)
	if err != nil {
		log.Printf("ERROR: unable to delete oa work %s: %s", workID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) deleteWork(namespace string, id string) error {
	log.Printf("INFO: get %s work %s for deletion", namespace, id)
	delObj, err := svc.EasyStore.GetByKey(namespace, id, uvaeasystore.BaseComponent)
	if err != nil {
		return fmt.Errorf("unable to get %s work %s: %s", namespace, id, err.Error())
	}

	log.Printf("INFO: delete %s work %s", namespace, id)
	_, err = svc.EasyStore.Delete(delObj, uvaeasystore.AllComponents)
	return err
}

func (svc *serviceContext) oaUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update oa work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}

func (svc *serviceContext) etdUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update etd work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}
