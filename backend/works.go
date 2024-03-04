package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) deleteWork(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to delete work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}

func (svc *serviceContext) oaUpdate(c *gin.Context) {
	workID := c.Param("id")
	log.Printf("INFO: request to update work %s", workID)

	c.String(http.StatusNotImplemented, "not implemented")
}
