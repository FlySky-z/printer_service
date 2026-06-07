package handler

import (
	"printer/services"

	"github.com/gin-gonic/gin"
)

type SoftwareStatus struct {
	WPS     bool `json:"wps"`
	Acrobat bool `json:"acrobat"`
}

var cachedSoftwareStatus SoftwareStatus

func InitSoftwareStatus() {
	cachedSoftwareStatus = SoftwareStatus{
		WPS:     services.CheckWPS(),
		Acrobat: services.CheckAcrobat(),
	}
}

func HandleSoftwareStatus(c *gin.Context) {
	c.JSON(200, cachedSoftwareStatus)
}

func HandlePrinterList(c *gin.Context) {
	printers, err := services.QueryPrinters()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	names := make([]string, 0, len(printers))
	for _, p := range printers {
		names = append(names, p.Name)
	}
	c.JSON(200, names)
}

func HandlePrinterJobs(c *gin.Context) {
	name := c.Param("name")
	jobs, err := services.QueryPrinterJobs(name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, jobs)
}

func HandlePrinterStatus(c *gin.Context) {
	printers, err := services.QueryPrinters()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, printers)
}
