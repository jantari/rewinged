package controllers

import (
    "github.com/gin-gonic/gin"

    "rewinged/models"
)

func GetInformation(c *gin.Context) {
    response := new(models.API_Information_1_1_0)
    response.Data.SourceIdentifier = "rewinged"
    response.Data.ServerSupportedVersions = []string{"1.1.0"}
    c.JSON(200, response)
}

