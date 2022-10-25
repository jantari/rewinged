package controllers

import (
    "github.com/gin-gonic/gin"

    "rewinged/models"
)

func GetInformation(c *gin.Context) {
    response := new(models.Information)
    response.Data.SourceIdentifier = "rewinged"
    response.Data.ServerSupportedVersions = []string{"1.1.0"}
    c.JSON(200, response)
}

